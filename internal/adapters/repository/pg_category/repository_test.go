package pg_category

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/category"
	"github.com/amirbakhtiari/abrcoffee/internal/core/ports"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupCategoryRepositoryTest(t *testing.T) (ports.CategoryRepository, *gorm.DB) {
	t.Helper()

	dsn := "host=127.0.0.1 user=abrcoffee_test password=test_only_password dbname=abrcoffee_test port=5434 sslmode=disable"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql.DB: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("ping test database: %v", err)
	}

	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Errorf("close test database: %v", err)
		}
	})

	// Safety guard: never run repository tests against another database.
	var databaseName string
	if err := db.Raw("SELECT current_database()").Scan(&databaseName).Error; err != nil {
		t.Fatalf("get current database: %v", err)
	}

	if databaseName != "abrcoffee_test" {
		t.Fatalf("unsafe test database %q; expected abrcoffee_test", databaseName)
	}

	if err := db.AutoMigrate(&CategoryEntity{}); err != nil {
		t.Fatalf("migrate category table: %v", err)
	}

	// Ensure the test database uses the same partial unique indexes
	// expected by the production repository.
	statements := []string{
		`DROP INDEX IF EXISTS idx_categories_name`,
		`DROP INDEX IF EXISTS idx_categories_slug`,
		`CREATE UNIQUE INDEX idx_categories_name ON categories (name) WHERE deleted_at IS NULL`,
		`CREATE UNIQUE INDEX idx_categories_slug ON categories (slug) WHERE deleted_at IS NULL`,
	}

	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("prepare category indexes: %v", err)
		}
	}

	// This cleanup is restricted to the explicitly verified test database.
	if err := db.Exec("DELETE FROM categories").Error; err != nil {
		t.Fatalf("clean category test data: %v", err)
	}

	return NewPostgresRepository(db), db
}

func newTestCategory(t *testing.T, name, slug string, priority int) *category.Category {
	t.Helper()

	cat, err := category.NewCategory(
		name,
		slug,
		true,
		priority,
		time.Now,
	)
	if err != nil {
		t.Fatalf("create domain category: %v", err)
	}

	return cat
}

func TestRepositoryCreateAndGet(t *testing.T) {
	repo, _ := setupCategoryRepositoryTest(t)
	ctx := context.Background()

	created, err := repo.Create(
		ctx,
		newTestCategory(t, "Repo Test Coffee", "repo-test-coffee", 10),
	)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if created.ID() == 0 {
		t.Fatal("Create() returned category with zero ID")
	}

	if created.Name() != "Repo Test Coffee" {
		t.Errorf("Name() = %q, want %q", created.Name(), "Repo Test Coffee")
	}

	gotByID, err := repo.GetByID(ctx, created.ID())
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if gotByID.ID() != created.ID() {
		t.Errorf("GetByID() ID = %d, want %d", gotByID.ID(), created.ID())
	}

	gotBySlug, err := repo.GetBySlug(ctx, "repo-test-coffee")
	if err != nil {
		t.Fatalf("GetBySlug() error = %v", err)
	}

	if gotBySlug.ID() != created.ID() {
		t.Errorf("GetBySlug() ID = %d, want %d", gotBySlug.ID(), created.ID())
	}
}

func TestRepositoryGetNotFound(t *testing.T) {
	repo, _ := setupCategoryRepositoryTest(t)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, 999999)
	if !errors.Is(err, category.ErrCategoryNotFound) {
		t.Errorf("GetByID() error = %v, want ErrCategoryNotFound", err)
	}

	_, err = repo.GetBySlug(ctx, "repo-test-missing")
	if !errors.Is(err, category.ErrCategoryNotFound) {
		t.Errorf("GetBySlug() error = %v, want ErrCategoryNotFound", err)
	}
}

func TestRepositoryGetAllPaginationAndOrdering(t *testing.T) {
	repo, _ := setupCategoryRepositoryTest(t)
	ctx := context.Background()

	categories := []struct {
		name     string
		slug     string
		priority int
	}{
		{"Repo Test C", "repo-test-c", 30},
		{"Repo Test A", "repo-test-a", 10},
		{"Repo Test B", "repo-test-b", 20},
	}

	for _, item := range categories {
		if _, err := repo.Create(
			ctx,
			newTestCategory(t, item.name, item.slug, item.priority),
		); err != nil {
			t.Fatalf("Create(%q) error = %v", item.name, err)
		}
	}

	got, err := repo.GetAll(ctx, 0, 2)
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("GetAll() returned %d categories, want 2", len(got))
	}

	if got[0].Priority() != 10 || got[1].Priority() != 20 {
		t.Errorf(
			"GetAll() priorities = [%d, %d], want [10, 20]",
			got[0].Priority(),
			got[1].Priority(),
		)
	}

	got, err = repo.GetAll(ctx, 2, 2)
	if err != nil {
		t.Fatalf("GetAll() second page error = %v", err)
	}

	if len(got) != 1 {
		t.Errorf("second page returned %d categories, want 1", len(got))
	}
}

func TestRepositoryDeleteAndReuseNameAndSlug(t *testing.T) {
	repo, _ := setupCategoryRepositoryTest(t)
	ctx := context.Background()

	original, err := repo.Create(
		ctx,
		newTestCategory(t, "Repo Test Reusable", "repo-test-reusable", 1),
	)
	if err != nil {
		t.Fatalf("Create() original error = %v", err)
	}

	if err := repo.Delete(ctx, original.ID()); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = repo.GetByID(ctx, original.ID())
	if !errors.Is(err, category.ErrCategoryNotFound) {
		t.Fatalf("GetByID() after delete error = %v, want ErrCategoryNotFound", err)
	}

	recreated, err := repo.Create(
		ctx,
		newTestCategory(t, "Repo Test Reusable", "repo-test-reusable", 2),
	)
	if err != nil {
		t.Fatalf("Create() after soft delete error = %v", err)
	}

	if recreated.ID() == original.ID() {
		t.Error("recreated category should have a different ID")
	}

	if err := repo.Delete(ctx, 999999); !errors.Is(err, category.ErrCategoryNotFound) {
		t.Errorf("Delete() missing ID error = %v, want ErrCategoryNotFound", err)
	}
}

func TestRepositoryCreateDuplicateNameAndSlug(t *testing.T) {
	repo, _ := setupCategoryRepositoryTest(t)
	ctx := context.Background()

	_, err := repo.Create(
		ctx,
		newTestCategory(t, "Repo Test Duplicate", "repo-test-duplicate", 1),
	)
	if err != nil {
		t.Fatalf("Create() original error = %v", err)
	}

	t.Run("duplicate name", func(t *testing.T) {
		_, err := repo.Create(
			ctx,
			newTestCategory(t, "Repo Test Duplicate", "repo-test-another-slug", 2),
		)

		if !errors.Is(err, category.ErrNameCategoryAlreadyExists) {
			t.Errorf(
				"Create() error = %v, want ErrNameCategoryAlreadyExists",
				err,
			)
		}
	})

	t.Run("duplicate slug", func(t *testing.T) {
		_, err := repo.Create(
			ctx,
			newTestCategory(t, "Repo Test Another Name", "repo-test-duplicate", 3),
		)

		if !errors.Is(err, category.ErrSlugCategoryAlreadyExists) {
			t.Errorf(
				"Create() error = %v, want ErrSlugCategoryAlreadyExists",
				err,
			)
		}
	})
}

func TestRepositoryUpdate(t *testing.T) {
	repo, _ := setupCategoryRepositoryTest(t)
	ctx := context.Background()

	created, err := repo.Create(
		ctx,
		newTestCategory(t, "Repo Test Before Update", "repo-test-before-update", 10),
	)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	clock := func() time.Time {
		return time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	}

	if err := created.Edit("Repo Test After Update", "repo-test-after-update", 0, clock); err != nil {
		t.Fatalf("Edit() error = %v", err)
	}

	created.Deactivate(clock)

	if err := repo.Update(ctx, created); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	updated, err := repo.GetByID(ctx, created.ID())
	if err != nil {
		t.Fatalf("GetByID() after update error = %v", err)
	}

	if updated.Name() != "Repo Test After Update" {
		t.Errorf("Name() = %q, want %q", updated.Name(), "Repo Test After Update")
	}

	if updated.Slug() != "repo-test-after-update" {
		t.Errorf("Slug() = %q, want %q", updated.Slug(), "repo-test-after-update")
	}

	if updated.Priority() != 0 {
		t.Errorf("Priority() = %d, want 0", updated.Priority())
	}

	if updated.IsActive() {
		t.Error("IsActive() = true, want false")
	}
}

func TestRepositoryUpdateNotFound(t *testing.T) {
	repo, _ := setupCategoryRepositoryTest(t)
	ctx := context.Background()

	missing := category.ReconstructCategory(
		999999,
		"Repo Test Missing",
		"repo-test-missing",
		0,
		true,
		time.Now(),
		time.Now(),
	)

	err := repo.Update(ctx, missing)
	if !errors.Is(err, category.ErrCategoryNotFound) {
		t.Errorf("Update() error = %v, want ErrCategoryNotFound", err)
	}
}

func TestRepositoryUpdateDuplicateName(t *testing.T) {
	repo, _ := setupCategoryRepositoryTest(t)
	ctx := context.Background()

	first, err := repo.Create(
		ctx,
		newTestCategory(t, "Repo Test First", "repo-test-first", 1),
	)
	if err != nil {
		t.Fatalf("Create() first error = %v", err)
	}

	second, err := repo.Create(
		ctx,
		newTestCategory(t, "Repo Test Second", "repo-test-second", 2),
	)
	if err != nil {
		t.Fatalf("Create() second error = %v", err)
	}

	if err := second.Edit(
		first.Name(),
		"repo-test-second-updated",
		2,
		time.Now,
	); err != nil {
		t.Fatalf("Edit() error = %v", err)
	}

	err = repo.Update(ctx, second)
	if !errors.Is(err, category.ErrNameCategoryAlreadyExists) {
		t.Errorf(
			"Update() error = %v, want ErrNameCategoryAlreadyExists",
			err,
		)
	}
}

func TestRepositoryUpdateDuplicateSlug(t *testing.T) {
	repo, _ := setupCategoryRepositoryTest(t)
	ctx := context.Background()

	first, err := repo.Create(
		ctx,
		newTestCategory(t, "Repo Test Slug First", "repo-test-slug-first", 1),
	)
	if err != nil {
		t.Fatalf("Create() first error = %v", err)
	}

	second, err := repo.Create(
		ctx,
		newTestCategory(t, "Repo Test Slug Second", "repo-test-slug-second", 2),
	)
	if err != nil {
		t.Fatalf("Create() second error = %v", err)
	}

	if err := second.Edit(
		"Repo Test Slug Second",
		first.Slug(),
		2,
		time.Now,
	); err != nil {
		t.Fatalf("Edit() error = %v", err)
	}

	err = repo.Update(ctx, second)
	if !errors.Is(err, category.ErrSlugCategoryAlreadyExists) {
		t.Errorf(
			"Update() error = %v, want ErrSlugCategoryAlreadyExists",
			err,
		)
	}
}

/*func TestRepositoryCreateErrorMessagesAreTranslated(t *testing.T) {
	repo, _ := setupCategoryRepositoryTest(t)
	ctx := context.Background()

	_, err := repo.Create(
		ctx,
		newTestCategory(t, "Repo Test Translation", "repo-test-translation", 1),
	)
	if err != nil {
		t.Fatalf("Create() first error = %v", err)
	}

	_, err = repo.Create(
		ctx,
		newTestCategory(t, "Repo Test Translation", "repo-test-translation-2", 2),
	)

	if err == nil {
		t.Fatal("expected duplicate name error")
	}

	if !errors.Is(err, category.ErrNameCategoryAlreadyExists) {
		t.Fatalf("expected translated domain error, got %v", err)
	}

	if !errors.Is(err, fmt.Errorf("%w", category.ErrNameCategoryAlreadyExists)) {
		t.Fatalf("translated error does not preserve sentinel identity: %v", err)
	}
}
*/
