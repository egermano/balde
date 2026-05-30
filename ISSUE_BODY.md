## Problem

The current database migration system is very basic and will not scale as we add new features that require schema changes.

### Current State

The current migration system in both `store.go` and `encrypted_store.go` uses a simple `CREATE TABLE IF NOT EXISTS` approach:

```go
func (s *SQLiteStore) migrate() error {
    schema := `
    CREATE TABLE IF NOT EXISTS accounts (...);
    CREATE TABLE IF NOT EXISTS buckets (...);
    CREATE TABLE IF NOT EXISTS transactions (...);
    `
    _, err := s.db.Exec(schema)
    return err
}
```

### Issues with Current System

1. **No version tracking** - doesn't know which migrations have been applied
2. **No incremental migrations** - runs the same schema every time  
3. **No schema versioning** - can't detect when schema needs upgrading
4. **No data migration** - can't transform existing data when schema changes

### Future Schema Changes (Phase 1+)

**Phase 1 (Post-MVP):**
- Tag/Taxonomy system → Need `tags` table, `transaction_tags` junction table
- Rule-based auto-categorization → Need `rules` table
- Transfer between buckets → Need `transfers` table or modify transactions

**Phase 2:**
- Debt accounts & debt payment buckets → Need `debts` table
- Recurring transaction scheduling → Need `recurring_transactions` table
- Savings goals with target dates → Need `goals` table

**Phase 3:**
- Multi-currency budgets → Need `currencies` table, exchange rates
- Multi-user/shared budgets → Need `users`, `permissions` tables

## Possible Solutions

### Option 1: Simple Versioned Migrations (Recommended for MVP Evolution)

```go
// store/migrations.go
type Migration struct {
    Version int
    Name    string
    Up      func(*sql.DB) error
    Down    func(*sql.DB) error
}

var migrations = []Migration{
    {Version: 1, Name: "initial_schema", Up: createInitialTables, Down: dropAllTables},
    {Version: 2, Name: "tags_system", Up: addTagsSystem, Down: removeTagsSystem},
    {Version: 3, Name: "recurring_transactions", Up: addRecurringTransactions, Down: removeRecurringTransactions},
}

// store/store.go
func (s *SQLiteStore) migrate() error {
    currentVersion, err := s.getSchemaVersion()
    if err != nil && !errors.Is(err, sql.ErrNoRows) {
        return err
    }
    
    for i := currentVersion; i < len(migrations); i++ {
        if err := migrations[i].Up(s.db); err != nil {
            return fmt.Errorf("migration %d failed: %w", migrations[i].Version, err)
        }
        if err := s.setSchemaVersion(migrations[i].Version); err != nil {
            return fmt.Errorf("set version %d failed: %w", migrations[i].Version, err)
        }
    }
    return nil
}
```

### Option 2: GORM/Ent Integration (For Future Major Features)

Use an existing ORM that handles migrations automatically:

```go
// Using Ent
func (s *SQLiteStore) migrate() error {
    ctx := context.Background()
    return client.Schema.Create(ctx, migrate.WithDropIndex(true), migrate.WithDropColumn(true))
}
```

### Option 3: Database Branching (Advanced)

Multiple database versions running side-by-side with conversion utilities.

### Option 4: External Migration Tooling

Use existing Go migration tools like:
- `atruso/go-sql-migrate`
- `pressly/goose`
- `golang-migrate/migrate`

## Implementation Plan

### Phase 1: Preparation (Before needing schema changes)
1. Add `schema_versions` table to track current version
2. Create migration structure and registry
3. Test migration system with dummy migrations
4. Document migration patterns and testing strategy

### Phase 2: First Schema Change (When adding tags)
1. Implement actual migrations for tag system
2. Add backward compatibility checks
3. Create data migration utilities
4. Test upgrade/downgrade paths

### Phase 3: Scale (Multiple migrations)
1. Create migration testing framework
2. Add migration rollback capabilities
3. Document migration best practices
4. Create migration examples and templates

## Benefits of Proper Migration System

1. **Schema Evolution**: Can safely add new features without data loss
2. **Version Control**: Track database schema changes alongside code changes
3. **Testing**: Ensure migrations work across different database states
4. **Rollback**: Safe rollback to previous versions if issues occur
5. **Documentation**: Clear audit trail of all schema changes
6. **Collaboration**: Multiple developers can work on schema changes safely

## Current Status

✅ **Current system is sufficient for MVP**
- No schema changes needed yet
- Simple initialization works fine
- No data migration needs currently

🚀 **Will need proper migration system for Phase 1+**
- Plan ahead to avoid technical debt
- Start with simple versioned approach
- Scale complexity as needed

## Acceptance Criteria

- [ ] Schema versioning system in place
- [ ] Migration registry and runner implemented
- [ ] Backup and rollback capabilities tested
- [ ] Documentation created for migration process
- [ ] First real migration (tags system) working correctly