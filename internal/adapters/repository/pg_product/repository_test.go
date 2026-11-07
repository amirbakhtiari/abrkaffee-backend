package pg_product

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/amirbakhtiari/abrcoffee/internal/adapters/repository/pg_category"
	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/product"
	"github.com/amirbakhtiari/abrcoffee/internal/core/ports"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupProductRepositoryTest(
	t *testing.T,
) (ports.ProductRepository, *gorm.DB, string) {
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

	var databaseName string
	if err := db.Raw("SELECT current_database()").Scan(&databaseName).Error; err != nil {
		t.Fatalf("get current database: %v", err)
	}

	if databaseName != "abrcoffee_test" {
		t.Fatalf("unsafe test database %q; expected abrcoffee_test", databaseName)
	}

	if err := db.AutoMigrate(
		&pg_category.CategoryEntity{},
		&ProductEntity{},
		&ProductCategoryEntity{},
	); err != nil {
		t.Fatalf("migrate product tables: %v", err)
	}

	// Ensure the test database uses the same partial unique indexes.
	statements := []string{
		`DROP INDEX IF EXISTS idx_products_code`,
		`DROP INDEX IF EXISTS idx_products_sku`,
		`DROP INDEX IF EXISTS idx_products_barcode`,
		`CREATE UNIQUE INDEX idx_products_code ON products (code) WHERE deleted_at IS NULL`,
		`CREATE UNIQUE INDEX idx_products_sku ON products (sku) WHERE deleted_at IS NULL AND sku <> ''`,
		`CREATE UNIQUE INDEX idx_products_barcode ON products (barcode) WHERE deleted_at IS NULL AND barcode <> ''`,
	}

	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("prepare product indexes: %v", err)
		}
	}

	prefix := fmt.Sprintf("it-product-%d-", time.Now().UnixNano())

	// Clean up only records created by this test setup.
	t.Cleanup(func() {
		if err := db.Exec(
			`DELETE FROM product_categories
			 WHERE product_id IN (
				SELECT id FROM products WHERE code LIKE ?
			 )`,
			prefix+"%",
		).Error; err != nil {
			t.Errorf("clean product relations: %v", err)
		}

		if err := db.Unscoped().
			Where("code LIKE ?", prefix+"%").
			Delete(&ProductEntity{}).Error; err != nil {
			t.Errorf("clean products: %v", err)
		}

		if err := db.Unscoped().
			Where("slug LIKE ?", prefix+"%").
			Delete(&pg_category.CategoryEntity{}).Error; err != nil {
			t.Errorf("clean categories: %v", err)
		}
	})

	return NewPostgresRepository(db), db, prefix
}

func createTestCategory(
	t *testing.T,
	db *gorm.DB,
	prefix string,
	name string,
) uint64 {
	t.Helper()

	entity := pg_category.CategoryEntity{
		Name:     prefix + name,
		Slug:     prefix + name,
		IsActive: true,
		Priority: 1,
	}

	if err := db.Create(&entity).Error; err != nil {
		t.Fatalf("create test category: %v", err)
	}

	return entity.ID
}

func newTestProduct(
	t *testing.T,
	prefix string,
	categoryIDs []uint64,
) *product.Product {
	t.Helper()

	item, err := product.NewProduct(
		"Integration Test Product",
		prefix+"code",
		prefix+"sku",
		prefix+"barcode",
		"Integration test description",
		true,
		categoryIDs,
		product.DefaultClock,
	)
	if err != nil {
		t.Fatalf("create domain product: %v", err)
	}

	return item
}

func TestRepositoryCreateAndGetByID(t *testing.T) {
	repo, db, prefix := setupProductRepositoryTest(t)
	ctx := context.Background()

	categoryA := createTestCategory(t, db, prefix, "category-a")
	categoryB := createTestCategory(t, db, prefix, "category-b")

	item := newTestProduct(t, prefix, []uint64{
		categoryA,
		categoryB,
	})

	created, err := repo.Create(ctx, item)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if created.ID() == 0 {
		t.Fatal("Create() returned product with zero ID")
	}

	if created.Code() != item.Code() {
		t.Errorf("Code() = %q, want %q", created.Code(), item.Code())
	}

	if created.Name() != item.Name() {
		t.Errorf("Name() = %q, want %q", created.Name(), item.Name())
	}

	got, err := repo.GetByID(ctx, created.ID())
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.ID() != created.ID() {
		t.Errorf("GetByID() ID = %d, want %d", got.ID(), created.ID())
	}

	if got.Code() != item.Code() {
		t.Errorf("GetByID() Code = %q, want %q", got.Code(), item.Code())
	}

	if got.SKU() != item.SKU() {
		t.Errorf("GetByID() SKU = %q, want %q", got.SKU(), item.SKU())
	}

	if got.Barcode() != item.Barcode() {
		t.Errorf("GetByID() Barcode = %q, want %q", got.Barcode(), item.Barcode())
	}

	if len(got.CategoryIDs()) != 2 {
		t.Fatalf(
			"GetByID() returned %d category IDs, want 2",
			len(got.CategoryIDs()),
		)
	}

	gotIDs := got.CategoryIDs()

	if gotIDs[0] != categoryA || gotIDs[1] != categoryB {
		t.Errorf(
			"CategoryIDs() = %v, want [%d %d]",
			gotIDs,
			categoryA,
			categoryB,
		)
	}
}

func TestRepositoryGetNotFound(t *testing.T) {
	repo, _, _ := setupProductRepositoryTest(t)

	_, err := repo.GetByID(context.Background(), 999999999)

	if !errors.Is(err, product.ErrProductNotFound) {
		t.Errorf(
			"GetByID() error = %v, want ErrProductNotFound",
			err,
		)
	}
}

func TestRepositoryCreateRollbackOnInvalidCategory(t *testing.T) {
	repo, db, prefix := setupProductRepositoryTest(t)
	ctx := context.Background()

	item := newTestProduct(t, prefix, []uint64{999999999})

	_, err := repo.Create(ctx, item)
	if err == nil {
		t.Fatal("Create() expected error for invalid category ID")
	}

	var count int64
	if err := db.Model(&ProductEntity{}).
		Where("code = ?", item.Code()).
		Count(&count).Error; err != nil {
		t.Fatalf("count products: %v", err)
	}

	if count != 0 {
		t.Errorf(
			"product count after failed transaction = %d, want 0",
			count,
		)
	}
}

func TestRepositoryUpdate(t *testing.T) {
	repo, db, prefix := setupProductRepositoryTest(t)
	ctx := context.Background()

	categoryA := createTestCategory(t, db, prefix, "cat-update-a")
	categoryB := createTestCategory(t, db, prefix, "cat-update-b")

	item := newTestProduct(t, prefix, []uint64{categoryA, categoryB})

	created, err := repo.Create(ctx, item)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// ایجاد دسته‌بندی جدید درست قبل از ویرایش برای اطمینان از معتبر بودن ID
	categoryC := createTestCategory(t, db, prefix, "cat-update-c")

	err = created.Edit(
		"Updated Product",
		prefix+"updated-code",
		prefix+"updated-sku",
		prefix+"updated-barcode",
		"Updated description",
		[]uint64{categoryC},
		product.DefaultClock,
	)
	if err != nil {
		t.Fatalf("Edit() error = %v", err)
	}

	updated, err := repo.Update(ctx, created)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if updated.Name() != "Updated Product" {
		t.Errorf("Name() = %q, want Updated Product", updated.Name())
	}

	if updated.Code() != prefix+"updated-code" {
		t.Errorf("Code() = %q, want %q", updated.Code(), prefix+"updated-code")
	}

	if updated.Description() != "Updated description" {
		t.Errorf("Description() = %q, want Updated description", updated.Description())
	}

	categoryIDs := updated.CategoryIDs()
	if len(categoryIDs) != 1 || categoryIDs[0] != categoryC {
		t.Errorf("CategoryIDs() = %v, want [%d]", categoryIDs, categoryC)
	}

	persisted, err := repo.GetByID(ctx, created.ID())
	if err != nil {
		t.Fatalf("GetByID() after Update error = %v", err)
	}

	if persisted.Code() != prefix+"updated-code" {
		t.Errorf("persisted Code() = %q, want %q", persisted.Code(), prefix+"updated-code")
	}

	persistedIDs := persisted.CategoryIDs()
	if len(persistedIDs) != 1 || persistedIDs[0] != categoryC {
		t.Errorf("persisted CategoryIDs() = %v, want [%d]", persistedIDs, categoryC)
	}
}

func TestRepositoryUpdateNotFound(t *testing.T) {
	repo, _, prefix := setupProductRepositoryTest(t)

	item := product.ReconstructProduct(
		999999999,
		"Missing Product",
		prefix+"missing-code",
		"",
		"",
		"",
		true,
		nil,
		time.Now(),
		time.Now(),
	)

	_, err := repo.Update(context.Background(), item)

	if !errors.Is(err, product.ErrProductNotFound) {
		t.Errorf("Update() error = %v, want ErrProductNotFound", err)
	}
}

func TestRepositoryUpdateRollbackOnInvalidCategory(t *testing.T) {
	repo, db, prefix := setupProductRepositoryTest(t)
	ctx := context.Background()

	categoryA := createTestCategory(t, db, prefix, "category-a")
	item := newTestProduct(t, prefix, []uint64{categoryA})

	created, err := repo.Create(ctx, item)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	oldCode := created.Code()

	err = created.Edit(
		"Updated Product",
		prefix+"updated-code",
		prefix+"updated-sku",
		prefix+"updated-barcode",
		"Updated description",
		[]uint64{999999999},
		product.DefaultClock,
	)
	if err != nil {
		t.Fatalf("Edit() error = %v", err)
	}

	p1, err := repo.Update(ctx, created)
	fmt.Println(p1)
	if err == nil {
		t.Fatal("Update() expected error for invalid category ID")
	}

	persisted, err := repo.GetByID(ctx, created.ID())
	if err != nil {
		t.Fatalf("GetByID() after failed Update error = %v", err)
	}

	if persisted.Code() != oldCode {
		t.Errorf("Code() after failed Update = %q, want %q", persisted.Code(), oldCode)
	}

	if len(persisted.CategoryIDs()) != 1 || persisted.CategoryIDs()[0] != categoryA {
		t.Errorf("CategoryIDs() after failed Update = %v, want [%d]", persisted.CategoryIDs(), categoryA)
	}
}
func TestRepositoryList_PaginationAndSorting(t *testing.T) {
	repo, _, prefix := setupProductRepositoryTest(t)
	ctx := context.Background()

	// ایجاد ۳ محصول با کد و نام مشخص
	names := []string{"Product A", "Product B", "Product C"}
	for i, name := range names {
		item, err := product.NewProduct(
			name,
			fmt.Sprintf("%scode-%d", prefix, i+1),
			fmt.Sprintf("%ssku-%d", prefix, i+1),
			fmt.Sprintf("%sbarcode-%d", prefix, i+1),
			"Desc",
			true,
			nil,
			product.DefaultClock,
		)
		if err != nil {
			t.Fatalf("create product %d: %v", i+1, err)
		}
		if _, err := repo.Create(ctx, item); err != nil {
			t.Fatalf("save product %d: %v", i+1, err)
		}
	}

	// ۱. تست صفحه اول با PageSize = 2
	items, total, err := repo.List(ctx, ports.ProductListQuery{
		Search:   prefix,
		Page:     1,
		PageSize: 2,
	})
	if err != nil {
		t.Fatalf("List() page 1 error = %v", err)
	}

	if total != 3 {
		t.Errorf("List() page 1 total = %d, want 3", total)
	}
	if len(items) != 2 {
		t.Fatalf("List() page 1 items len = %d, want 2", len(items))
	}

	// مرتب‌سازی بر اساس created_at DESC, id DESC است؛ پس آخرین محصول (شماره ۳) باید اولین رکورد باشد
	wantFirstCode := fmt.Sprintf("%scode-3", prefix)
	if items[0].Code() != wantFirstCode {
		t.Errorf("List() page 1 items[0].Code() = %q, want %q", items[0].Code(), wantFirstCode)
	}

	// ۲. تست صفحه دوم با PageSize = 2 (باید ۱ آیتم برگردد)
	itemsPage2, totalPage2, err := repo.List(ctx, ports.ProductListQuery{
		Search:   prefix,
		Page:     2,
		PageSize: 2,
	})
	if err != nil {
		t.Fatalf("List() page 2 error = %v", err)
	}

	if totalPage2 != 3 {
		t.Errorf("List() page 2 total = %d, want 3", totalPage2)
	}
	if len(itemsPage2) != 1 {
		t.Fatalf("List() page 2 items len = %d, want 1", len(itemsPage2))
	}

	wantSecondPageCode := fmt.Sprintf("%scode-1", prefix)
	if itemsPage2[0].Code() != wantSecondPageCode {
		t.Errorf("List() page 2 items[0].Code() = %q, want %q", itemsPage2[0].Code(), wantSecondPageCode)
	}

	// ۳. تست صفحه خالی (Page = 3)
	itemsPage3, totalPage3, err := repo.List(ctx, ports.ProductListQuery{
		Search:   prefix,
		Page:     3,
		PageSize: 2,
	})
	if err != nil {
		t.Fatalf("List() page 3 error = %v", err)
	}

	if totalPage3 != 3 {
		t.Errorf("List() page 3 total = %d, want 3", totalPage3)
	}
	if len(itemsPage3) != 0 {
		t.Errorf("List() page 3 items len = %d, want 0", len(itemsPage3))
	}
}

func TestRepositoryList_FiltersAndSearch(t *testing.T) {
	repo, db, prefix := setupProductRepositoryTest(t)
	ctx := context.Background()

	catA := createTestCategory(t, db, prefix, "cat-a")
	catB := createTestCategory(t, db, prefix, "cat-b")

	// محصول ۱: فعال، دسته‌بندی A، نام و کد خاص
	p1, err := product.NewProduct(
		"Special Espresso",
		prefix+"code-esp",
		prefix+"sku-esp",
		prefix+"bar-111",
		"Dark roast",
		true,
		[]uint64{catA},
		product.DefaultClock,
	)
	if err != nil {
		t.Fatalf("create p1: %v", err)
	}
	if _, err := repo.Create(ctx, p1); err != nil {
		t.Fatalf("save p1: %v", err)
	}

	// محصول ۲: غیرفعال، دسته‌بندی B
	p2, err := product.NewProduct(
		"Filter Colombia",
		prefix+"code-flt",
		prefix+"sku-flt",
		prefix+"bar-222",
		"Light roast",
		false,
		[]uint64{catB},
		product.DefaultClock,
	)
	if err != nil {
		t.Fatalf("create p2: %v", err)
	}
	if _, err := repo.Create(ctx, p2); err != nil {
		t.Fatalf("save p2: %v", err)
	}

	// ۱. تست فیلتر IsActive = true
	activeTrue := true
	activeItems, totalActive, err := repo.List(ctx, ports.ProductListQuery{
		Search:   prefix,
		IsActive: &activeTrue,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("List() active error = %v", err)
	}
	if totalActive != 1 || len(activeItems) != 1 {
		t.Fatalf("List() active total = %d, items len = %d; want 1", totalActive, len(activeItems))
	}
	if activeItems[0].Code() != prefix+"code-esp" {
		t.Errorf("List() active product code = %q, want %q", activeItems[0].Code(), prefix+"code-esp")
	}

	// ۲. تست فیلتر CategoryID
	catItems, totalCat, err := repo.List(ctx, ports.ProductListQuery{
		Search:     prefix,
		CategoryID: catB,
		Page:       1,
		PageSize:   10,
	})
	if err != nil {
		t.Fatalf("List() category filter error = %v", err)
	}
	if totalCat != 1 || len(catItems) != 1 {
		t.Fatalf("List() category filter total = %d, items len = %d; want 1", totalCat, len(catItems))
	}
	if catItems[0].Code() != prefix+"code-flt" {
		t.Errorf("List() category filter code = %q, want %q", catItems[0].Code(), prefix+"code-flt")
	}

	// ۳. تست جستجوی متنی روی Name (به‌صورت Case-insensitive)
	searchItems, totalSearch, err := repo.List(ctx, ports.ProductListQuery{
		Search:   "espresso",
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("List() search name error = %v", err)
	}
	if totalSearch != 1 || len(searchItems) != 1 {
		t.Fatalf("List() search name total = %d, items len = %d; want 1", totalSearch, len(searchItems))
	}
	if searchItems[0].Code() != prefix+"code-esp" {
		t.Errorf("List() search name code = %q, want %q", searchItems[0].Code(), prefix+"code-esp")
	}

	// ۴. تست جستجوی متنی روی Barcode
	barItems, totalBar, err := repo.List(ctx, ports.ProductListQuery{
		Search:   prefix + "bar-222",
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("List() search barcode error = %v", err)
	}
	if totalBar != 1 || len(barItems) != 1 {
		t.Fatalf("List() search barcode total = %d, items len = %d; want 1", totalBar, len(barItems))
	}
	if barItems[0].Code() != prefix+"code-flt" {
		t.Errorf("List() search barcode code = %q, want %q", barItems[0].Code(), prefix+"code-flt")
	}
}

func TestRepositoryDelete(t *testing.T) {
	repo, db, prefix := setupProductRepositoryTest(t)
	ctx := context.Background()

	catA := createTestCategory(t, db, prefix, "cat-del")
	item := newTestProduct(t, prefix, []uint64{catA})

	created, err := repo.Create(ctx, item)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// ۱. حذف موفق محصول
	if err := repo.Delete(ctx, created.ID()); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// ۲. اطمینان از اینکه GetByID دیگر محصول را برنمی‌گرداند
	_, err = repo.GetByID(ctx, created.ID())
	if !errors.Is(err, product.ErrProductNotFound) {
		t.Errorf("GetByID() after Delete error = %v, want ErrProductNotFound", err)
	}

	// ۳. اطمینان از اینکه در List ظاهر نمی‌شود
	items, total, err := repo.List(ctx, ports.ProductListQuery{
		Search: prefix,
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if total != 0 || len(items) != 0 {
		t.Errorf("List() total = %d, items len = %d; want 0", total, len(items))
	}

	// ۴. اطمینان از پاک شدن روابط در جدول واسط product_categories
	var relCount int64
	if err := db.Model(&ProductCategoryEntity{}).
		Where("product_id = ?", created.ID()).
		Count(&relCount).Error; err != nil {
		t.Fatalf("count relations: %v", err)
	}
	if relCount != 0 {
		t.Errorf("product_categories count after delete = %d, want 0", relCount)
	}

	// ۵. بررسی سطح دیتابیس: رکورد در products به شکل Soft Delete باقی مانده است
	var entity ProductEntity
	err = db.Unscoped().First(&entity, created.ID()).Error
	if err != nil {
		t.Fatalf("Unscoped().First() error = %v", err)
	}
	if !entity.DeletedAt.Valid {
		t.Error("expected DeletedAt to be set for soft deleted product")
	}
}

func TestRepositoryDeleteNotFound(t *testing.T) {
	repo, _, _ := setupProductRepositoryTest(t)

	err := repo.Delete(context.Background(), 999999999)
	if !errors.Is(err, product.ErrProductNotFound) {
		t.Errorf("Delete() non-existent product error = %v, want ErrProductNotFound", err)
	}
}

func TestRepositoryUniqueConstraints(t *testing.T) {
	repo, _, prefix := setupProductRepositoryTest(t)
	ctx := context.Background()

	// ۱. محصول پایه
	base, err := product.NewProduct(
		"Base Product",
		prefix+"code-unique",
		prefix+"sku-unique",
		prefix+"bar-unique",
		"Desc",
		true,
		nil,
		product.DefaultClock,
	)
	if err != nil {
		t.Fatalf("create base product: %v", err)
	}
	if _, err := repo.Create(ctx, base); err != nil {
		t.Fatalf("save base product: %v", err)
	}

	// ۲. تکرار Code یکسان -> باید ErrProductCodeAlreadyExists بدهد
	dupCode, err := product.NewProduct(
		"Duplicate Code",
		prefix+"code-unique",
		prefix+"sku-other-1",
		prefix+"bar-other-1",
		"Desc",
		true,
		nil,
		product.DefaultClock,
	)
	if err != nil {
		t.Fatalf("create dupCode: %v", err)
	}
	_, err = repo.Create(ctx, dupCode)
	if !errors.Is(err, product.ErrProductCodeAlreadyExists) {
		t.Errorf("Create() dupCode error = %v, want ErrProductCodeAlreadyExists", err)
	}

	// ۳. تکرار SKU یکسان -> باید ErrProductSKUAlreadyExists بدهد
	dupSKU, err := product.NewProduct(
		"Duplicate SKU",
		prefix+"code-other-1",
		prefix+"sku-unique",
		prefix+"bar-other-2",
		"Desc",
		true,
		nil,
		product.DefaultClock,
	)
	if err != nil {
		t.Fatalf("create dupSKU: %v", err)
	}
	_, err = repo.Create(ctx, dupSKU)
	if !errors.Is(err, product.ErrProductSKUAlreadyExists) {
		t.Errorf("Create() dupSKU error = %v, want ErrProductSKUAlreadyExists", err)
	}

	// ۴. تکرار Barcode یکسان -> باید ErrProductBarcodeAlreadyExists بدهد
	dupBar, err := product.NewProduct(
		"Duplicate Barcode",
		prefix+"code-other-2",
		prefix+"sku-other-2",
		prefix+"bar-unique",
		"Desc",
		true,
		nil,
		product.DefaultClock,
	)
	if err != nil {
		t.Fatalf("create dupBar: %v", err)
	}
	_, err = repo.Create(ctx, dupBar)
	if !errors.Is(err, product.ErrProductBarcodeAlreadyExists) {
		t.Errorf("Create() dupBar error = %v, want ErrProductBarcodeAlreadyExists", err)
	}

	// ۵. مجاز بودن چندین مقدار خالی برای SKU و Barcode
	empty1, err := product.NewProduct(
		"Empty Identifiers 1",
		prefix+"code-empty-1",
		"",
		"",
		"Desc",
		true,
		nil,
		product.DefaultClock,
	)
	if err != nil {
		t.Fatalf("create empty1: %v", err)
	}
	if _, err := repo.Create(ctx, empty1); err != nil {
		t.Errorf("Create() empty1 error = %v, want success", err)
	}

	empty2, err := product.NewProduct(
		"Empty Identifiers 2",
		prefix+"code-empty-2",
		"",
		"",
		"Desc",
		true,
		nil,
		product.DefaultClock,
	)
	if err != nil {
		t.Fatalf("create empty2: %v", err)
	}
	if _, err := repo.Create(ctx, empty2); err != nil {
		t.Errorf("Create() empty2 error = %v, want success (multiple empty SKU/Barcode allowed)", err)
	}
}

func TestRepositoryUniqueConstraints_ReuseAfterSoftDelete(t *testing.T) {
	repo, _, prefix := setupProductRepositoryTest(t)
	ctx := context.Background()

	code := prefix + "reuse-code"
	sku := prefix + "reuse-sku"
	barcode := prefix + "reuse-barcode"

	p1, err := product.NewProduct(
		"Product To Delete",
		code,
		sku,
		barcode,
		"Desc",
		true,
		nil,
		product.DefaultClock,
	)
	if err != nil {
		t.Fatalf("create p1: %v", err)
	}
	created, err := repo.Create(ctx, p1)
	if err != nil {
		t.Fatalf("save p1: %v", err)
	}

	// حذف محصول اول
	if err := repo.Delete(ctx, created.ID()); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// ایجاد محصول دوم با همان Code، SKU و Barcode دقیقاً
	p2, err := product.NewProduct(
		"Reused Product",
		code,
		sku,
		barcode,
		"New Desc",
		true,
		nil,
		product.DefaultClock,
	)
	if err != nil {
		t.Fatalf("create p2: %v", err)
	}

	// این عملیات به خاطر وجود Partial Index باید با موفقیت ثبت شود
	reused, err := repo.Create(ctx, p2)
	if err != nil {
		t.Fatalf("Create() reusing identifiers after soft delete failed: %v", err)
	}

	if reused.ID() == 0 || reused.ID() == created.ID() {
		t.Errorf("expected new unique ID for reused product, got %d", reused.ID())
	}
}
func TestRepositoryUpdateStatus(t *testing.T) {
	repo, db, prefix := setupProductRepositoryTest(t)
	ctx := context.Background()

	item := newTestProduct(t, prefix, nil)

	created, err := repo.Create(ctx, item)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	updatedAt := time.Date(2026, time.November, 7, 10, 30, 0, 0, time.UTC)

	// غیرفعال‌کردن محصول
	err = repo.UpdateStatus(ctx, created.ID(), false, updatedAt)
	if err != nil {
		t.Fatalf("UpdateStatus(false) error = %v", err)
	}

	var entity ProductEntity
	if err := db.First(&entity, created.ID()).Error; err != nil {
		t.Fatalf("load product after deactivation: %v", err)
	}

	if entity.IsActive {
		t.Error("IsActive = true, want false")
	}

	if !entity.UpdatedAt.Equal(updatedAt) {
		t.Errorf("UpdatedAt = %v, want %v", entity.UpdatedAt, updatedAt)
	}

	// فعال‌کردن مجدد محصول
	nextUpdatedAt := updatedAt.Add(time.Minute)

	err = repo.UpdateStatus(ctx, created.ID(), true, nextUpdatedAt)
	if err != nil {
		t.Fatalf("UpdateStatus(true) error = %v", err)
	}

	if err := db.First(&entity, created.ID()).Error; err != nil {
		t.Fatalf("load product after activation: %v", err)
	}

	if !entity.IsActive {
		t.Error("IsActive = false, want true")
	}

	if !entity.UpdatedAt.Equal(nextUpdatedAt) {
		t.Errorf("UpdatedAt = %v, want %v", entity.UpdatedAt, nextUpdatedAt)
	}
}

func TestRepositoryUpdateStatusNotFound(t *testing.T) {
	repo, _, _ := setupProductRepositoryTest(t)

	err := repo.UpdateStatus(
		context.Background(),
		999999999,
		false,
		time.Now(),
	)

	if !errors.Is(err, product.ErrProductNotFound) {
		t.Errorf(
			"UpdateStatus() error = %v, want ErrProductNotFound",
			err,
		)
	}
}

func TestRepositoryUpdateStatusSoftDeleted(t *testing.T) {
	repo, db, prefix := setupProductRepositoryTest(t)
	ctx := context.Background()

	item := newTestProduct(t, prefix, nil)

	created, err := repo.Create(ctx, item)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := repo.Delete(ctx, created.ID()); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	err = repo.UpdateStatus(ctx, created.ID(), false, time.Now())
	if !errors.Is(err, product.ErrProductNotFound) {
		t.Errorf(
			"UpdateStatus() on soft-deleted product error = %v, want ErrProductNotFound",
			err,
		)
	}

	// بررسی اینکه وضعیت رکورد حذف‌شده تغییر نکرده است.
	var entity ProductEntity
	if err := db.Unscoped().First(&entity, created.ID()).Error; err != nil {
		t.Fatalf("load soft-deleted product: %v", err)
	}

	if !entity.IsActive {
		t.Error("soft-deleted product IsActive changed; want true")
	}
}
