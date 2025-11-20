package migrate

import (
	"database/sql"
	"fmt"

	m "github.com/go-sql-driver/mysql"

	migrate "github.com/rubenv/sql-migrate"

	"github.com/ppxb/oreo-admin-go/pkg/log"
)

func Do(options ...func(*Options)) error {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}

	if err := ensureDatabase(ops); err != nil {
		return err
	}

	db, err := sql.Open(ops.driver, ops.uri)
	if err != nil {
		log.WithContext(ops.ctx).WithError(err).Error("[DATABASE] Open %s(%s) failed", ops.driver, ops.uri)
		return err
	}
	defer db.Close()

	return withAdvisoryLock(ops, db, func() error {
		return executeMigration(ops, db)
	})
}

func ensureDatabase(ops *Options) error {
	cfg, err := m.ParseDSN(ops.uri)
	if err != nil {
		log.WithContext(ops.ctx).WithError(err).Error("[DATABASE] Invalid database uri")
		return err
	}

	dbName := cfg.DBName
	cfg.DBName = ""

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return err
	}
	defer db.Close()

	if _, err := db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", dbName)); err != nil {
		log.WithContext(ops.ctx).WithError(err).Error("[DATABASE] Create database failed")
		return err
	}
	return nil
}

func withAdvisoryLock(ops *Options, db *sql.DB, fn func() error) error {
	if err := waitForLock(ops, db); err != nil {
		return err
	}

	defer func() {
		if err := releaseLock(ops, db); err != nil {
			log.WithContext(ops.ctx).WithError(err).Error("[DATABASE] Release lock failed")
		}
	}()

	return fn()
}

func waitForLock(ops *Options, db *sql.DB) error {
	for {
		lockAcquired, err := acquireLock(ops, db)
		if err != nil {
			return err
		}
		if lockAcquired {
			log.WithContext(ops.ctx).Debug("[DATABASE] Advisory lock acquired")
			return nil
		}
		log.WithContext(ops.ctx).Debug("[DATABASE] Waiting for advisory lock...")
	}
}

func acquireLock(ops *Options, db *sql.DB) (bool, error) {
	var lockAcquired bool
	query := fmt.Sprintf("SELECT GET_LOCK('%s', 5)", ops.lockName)

	if err := db.QueryRow(query).Scan(&lockAcquired); err != nil {
		log.WithContext(ops.ctx).WithError(err).Error("[DATABASE] Acquire advisory lock failed")
		return false, err
	}
	return lockAcquired, nil
}

func releaseLock(ops *Options, db *sql.DB) error {
	query := fmt.Sprintf("SELECT RELEASE_LOCK('%s')", ops.lockName)
	if _, err := db.Exec(query); err != nil {
		log.WithContext(ops.ctx).WithError(err).Error("[DATABASE] Release advisory lock failed")
		return err
	}

	log.WithContext(ops.ctx).Debug("[DATABASE] Advisory lock released")
	return nil
}

func executeMigration(ops *Options, db *sql.DB) error {
	if ops.before != nil {
		if err := ops.before(ops.ctx); err != nil {
			log.WithContext(ops.ctx).WithError(err).Error("[DATABASE] Before callback failed")
			return err
		}
	}

	migrate.SetTable(ops.changeTable)
	source := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: ops.fs,
		Root:       ops.fsRoot,
	}

	if err := logMigrationStatus(ops, db, source); err != nil {
		return err
	}

	if _, err := migrate.Exec(db, ops.driver, source, migrate.Up); err != nil {
		log.WithContext(ops.ctx).WithError(err).Error("[DATABASE] Migration failed")
		return err
	}

	log.WithContext(ops.ctx).Info("[DATABASE] migration completed successfully")
	return nil
}

func logMigrationStatus(ops *Options, db *sql.DB, source *migrate.EmbedFileSystemMigrationSource) error {
	migrations, err := source.FindMigrations()
	if err != nil {
		log.WithContext(ops.ctx).WithError(err).Error("[DATABASE] Find migrations failed")
		return err
	}

	records, err := migrate.GetMigrationRecords(db, ops.driver)
	if err != nil {
		log.WithContext(ops.ctx).WithError(err).Error("[DATABASE] Find migration records failed")
		return err
	}

	pending, applied := categorizeMigrations(migrations, records)

	log.WithContext(ops.ctx).Debug("[DATABASE] Migration status: %d pending, %d applied", len(pending), len(applied))

	return nil
}

func categorizeMigrations(migrations []*migrate.Migration, records []*migrate.MigrationRecord) (pending, applied []string) {
	appliedSet := make(map[string]bool, len(records))
	for _, record := range records {
		appliedSet[record.Id] = true
	}

	for _, migration := range migrations {
		if appliedSet[migration.Id] {
			applied = append(applied, migration.Id)
		} else {
			pending = append(pending, migration.Id)
		}
	}
	return pending, applied
}
