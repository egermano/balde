package main

import (
	"fmt"
	"testing"
	"time"
)

type fakeVacationHost struct {
	buckets    []bucket
	allocations []struct {
		bucketID string
		amount   int64
	}
}

func (h *fakeVacationHost) call(method string, params any, output any) error {
	switch method {
	case "host/listBuckets":
		out := output.(*[]bucket)
		*out = append([]bucket(nil), h.buckets...)
		return nil
	case "host/addBucket":
		p := params.(map[string]any)
		created := bucket{ID: fmt.Sprintf("bkt-%d", len(h.buckets)+1), Name: p["name"].(string), Target: p["target"].(int64)}
		h.buckets = append(h.buckets, created)
		*(output.(*bucket)) = created
		return nil
	case "host/allocate":
		p := params.(map[string]any)
		h.allocations = append(h.allocations, struct {
			bucketID string
			amount   int64
		}{p["bucket_id"].(string), p["amount"].(int64)})
		return nil
	default:
		return fmt.Errorf("unexpected method %s", method)
	}
}

func TestPlanCreatesVacationBucketAndAllocatesFirstInstallment(t *testing.T) {
	host := &fakeVacationHost{}
	now := time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC)

	result, err := plan(host, []string{"Japan", "--date", "2027-03", "--budget", "1000"}, now)
	if err != nil {
		t.Fatalf("plan() error = %v, want nil", err)
	}

	if len(host.buckets) != 1 {
		t.Fatalf("created buckets = %d, want 1", len(host.buckets))
	}
	gotBucket := host.buckets[0]
	if gotBucket.Name != "vacation-japan-2027-03" || gotBucket.Target != 1000 {
		t.Errorf("bucket = %+v, want vacation-japan-2027-03 target 1000", gotBucket)
	}
	if len(host.allocations) != 1 || host.allocations[0].bucketID != "bkt-1" || host.allocations[0].amount != 200 {
		t.Errorf("allocations = %+v, want bkt-1 / 200", host.allocations)
	}

	output, ok := result["json"].(planOutput)
	if !ok {
		t.Fatalf("result json = %T, want planOutput", result["json"])
	}
	if output.Months != 5 || output.MonthlyInstallment != 200 || output.FinalInstallment != 200 || output.RemainingInstallments != 4 {
		t.Errorf("plan output = %+v", output)
	}
}

func TestPlanRoundsUpAndAdjustsFinalInstallment(t *testing.T) {
	host := &fakeVacationHost{}
	now := time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC)

	result, err := plan(host, []string{"Japan", "--date", "2027-01", "--budget", "1000"}, now)
	if err != nil {
		t.Fatal(err)
	}
	output := result["json"].(planOutput)
	if output.MonthlyInstallment != 334 || output.FinalInstallment != 332 {
		t.Errorf("output = %+v, want monthly 334 / final 332", output)
	}
}

func TestPlanRejectsExistingVacationBucketWithoutMutating(t *testing.T) {
	host := &fakeVacationHost{buckets: []bucket{{ID: "bkt-1", Name: "vacation-japan-2027-03", Target: 1000}}}
	now := time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC)

	_, err := plan(host, []string{"Japan", "--date", "2027-03", "--budget", "1000"}, now)
	if err == nil {
		t.Fatal("plan(existing) = nil error, want error")
	}
	if len(host.buckets) != 1 || len(host.allocations) != 0 {
		t.Errorf("host mutated after duplicate: buckets=%+v allocations=%+v", host.buckets, host.allocations)
	}
}

func TestScheduleReportsRemainingAllocation(t *testing.T) {
	host := &fakeVacationHost{buckets: []bucket{{ID: "bkt-1", Name: "vacation-japan-2027-03", Target: 1000, Balance: 200}}}
	now := time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC)

	result, err := schedule(host, []string{"Japan", "--date", "2027-03"}, now)
	if err != nil {
		t.Fatal(err)
	}
	json := result["json"].(map[string]any)
	if json["remaining"] != int64(800) || json["months"] != 4 || json["monthly_installment"] != int64(200) {
		t.Errorf("schedule json = %+v, want remaining 800 / 4 months / 200", json)
	}
}
