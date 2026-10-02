package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type bucket struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Target  int64  `json:"target"`
	Balance int64  `json:"balance"`
}

type planOutput struct {
	BucketName            string `json:"bucket_name"`
	BucketID              string `json:"bucket_id"`
	VacationDate          string `json:"vacation_date"`
	Budget                int64  `json:"budget"`
	Months                int    `json:"months"`
	MonthlyInstallment    int64  `json:"monthly_installment"`
	FinalInstallment      int64  `json:"final_installment"`
	AllocatedNow          int64  `json:"allocated_now"`
	RemainingInstallments int    `json:"remaining_installments"`
}

// vacationHost is the narrow mediated API the planner needs. The protocol
// client implements it; tests use an in-memory fake without a subprocess.
type vacationHost interface {
	call(method string, params any, output any) error
}

func main() {
	c := newClient(os.Stdin, os.Stdout)
	for {
		msg, err := c.read()
		if err != nil {
			return
		}
		switch msg.Method {
		case "initialize":
			_ = c.respond(msg.ID, map[string]any{
				"name": "vacation", "version": "0.1.0", "protocol": protocolVersion,
				"capabilities": []map[string]string{{"type": "command", "name": "vacation"}},
			})
		case "command/execute":
			var params struct {
				Args []string `json:"args"`
			}
			if err := json.Unmarshal(msg.Params, &params); err != nil {
				_ = c.respondError(msg.ID, -32602, "E_PARAMS: invalid command arguments")
				continue
			}
			result, err := execute(c, params.Args, time.Now())
			if err != nil {
				_ = c.respondError(msg.ID, 1, "E_VACATION: "+err.Error())
				continue
			}
			_ = c.respond(msg.ID, result)
		default:
			_ = c.respondError(msg.ID, -32601, "method not found: "+msg.Method)
		}
	}
}

func execute(c *client, args []string, now time.Time) (map[string]any, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("usage: vacation plan <destination> --date YYYY-MM --budget <cents>")
	}
	switch args[0] {
	case "plan":
		return plan(c, args[1:], now)
	case "schedule":
		return schedule(c, args[1:], now)
	default:
		return nil, fmt.Errorf("unknown command %q; use plan or schedule", args[0])
	}
}

func plan(c vacationHost, args []string, now time.Time) (map[string]any, error) {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return nil, fmt.Errorf("usage: vacation plan <destination> --date YYYY-MM --budget <cents>")
	}
	destination := args[0]
	flags := flag.NewFlagSet("vacation plan", flag.ContinueOnError)
	flags.SetOutput(ioDiscard{})
	date := flags.String("date", "", "")
	budgetText := flags.String("budget", "", "")
	if err := flags.Parse(args[1:]); err != nil {
		return nil, fmt.Errorf("usage: vacation plan <destination> --date YYYY-MM --budget <cents>")
	}
	budget, err := strconv.ParseInt(*budgetText, 10, 64)
	if err != nil || budget <= 0 {
		return nil, fmt.Errorf("budget must be a positive integer number of cents")
	}
	months, err := monthsUntilVacation(now, *date)
	if err != nil {
		return nil, err
	}
	bucketName := vacationBucketName(destination, *date)

	var buckets []bucket
	if err := c.call("host/listBuckets", map[string]any{}, &buckets); err != nil {
		return nil, err
	}
	for _, existing := range buckets {
		if existing.Name == bucketName {
			return nil, fmt.Errorf("vacation bucket %q already exists; use vacation schedule %q --date %s", bucketName, destination, *date)
		}
	}

	var created bucket
	if err := c.call("host/addBucket", map[string]any{"name": bucketName, "target": budget}, &created); err != nil {
		return nil, err
	}
	installment := monthlyInstallment(budget, months)
	if err := c.call("host/allocate", map[string]any{"bucket_id": created.ID, "amount": installment}, nil); err != nil {
		return nil, err
	}
	final := budget - installment*int64(months-1)
	output := planOutput{
		BucketName: bucketName, BucketID: created.ID, VacationDate: *date,
		Budget: budget, Months: months, MonthlyInstallment: installment,
		FinalInstallment: final, AllocatedNow: installment,
		RemainingInstallments: months - 1,
	}
	return planResult(output), nil
}

func schedule(c vacationHost, args []string, now time.Time) (map[string]any, error) {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return nil, fmt.Errorf("usage: vacation schedule <destination> --date YYYY-MM")
	}
	destination := args[0]
	flags := flag.NewFlagSet("vacation schedule", flag.ContinueOnError)
	flags.SetOutput(ioDiscard{})
	date := flags.String("date", "", "")
	if err := flags.Parse(args[1:]); err != nil {
		return nil, fmt.Errorf("usage: vacation schedule <destination> --date YYYY-MM")
	}
	months, err := monthsUntilVacation(now, *date)
	if err != nil {
		return nil, err
	}
	bucketName := vacationBucketName(destination, *date)
	var buckets []bucket
	if err := c.call("host/listBuckets", map[string]any{}, &buckets); err != nil {
		return nil, err
	}
	for _, existing := range buckets {
		if existing.Name != bucketName {
			continue
		}
		remaining := existing.Target - existing.Balance
		if remaining <= 0 {
			return map[string]any{
				"text": fmt.Sprintf("Vacation bucket %q is fully funded.", bucketName),
				"json": map[string]any{"bucket_name": bucketName, "remaining": int64(0), "fully_funded": true},
			}, nil
		}
		installment := monthlyInstallment(remaining, months)
		return map[string]any{
			"text": fmt.Sprintf("%d cents remain for %q. Allocate %d cents this month: balde allocate %d %s", remaining, bucketName, installment, installment, existing.ID),
			"json": map[string]any{"bucket_name": bucketName, "bucket_id": existing.ID, "remaining": remaining, "months": months, "monthly_installment": installment},
		}, nil
	}
	return nil, fmt.Errorf("vacation bucket %q not found", bucketName)
}

func planResult(output planOutput) map[string]any {
	text := fmt.Sprintf("Created %q. Allocated %d cents now. Save %d cents each month for %d more month(s); final installment: %d cents.", output.BucketName, output.AllocatedNow, output.MonthlyInstallment, output.RemainingInstallments, output.FinalInstallment)
	return map[string]any{"text": text, "json": output}
}

// ioDiscard keeps flag parsing errors out of the protocol stdout stream.
type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }
