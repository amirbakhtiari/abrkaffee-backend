package category

import (
	"regexp"
	"strings"
	"time"
)

const MaxCategoryNameLength = 100

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type Clock func() time.Time

var DefaultClock Clock = time.Now

type Category struct {
	id        uint64
	name      string
	slug      string
	isActive  bool
	priority  int
	createdAt time.Time
	updatedAt time.Time
}

func NewCategory(name, slug string, isActive bool, priority int, clock Clock) (*Category, error) {
	name = strings.TrimSpace(name)
	slug = normalizeSlug(slug)

	if err := validateFields(name, slug, priority); err != nil {
		return nil, err
	}

	now := clock()
	return &Category{
		name:      name,
		slug:      slug,
		isActive:  isActive,
		priority:  priority,
		createdAt: now,
		updatedAt: now,
	}, nil
}

func ReconstructCategory(
	id uint64,
	name, slug string,
	priority int,
	isActive bool,
	createdAt, updatedAt time.Time,
) *Category {
	return &Category{
		id:        id,
		name:      name,
		slug:      slug,
		isActive:  isActive,
		priority:  priority,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (c *Category) ID() uint64           { return c.id }
func (c *Category) Name() string         { return c.name }
func (c *Category) Slug() string         { return c.slug }
func (c *Category) IsActive() bool       { return c.isActive }
func (c *Category) Priority() int        { return c.priority }
func (c *Category) CreatedAt() time.Time { return c.createdAt }
func (c *Category) UpdatedAt() time.Time { return c.updatedAt }

func (c *Category) Edit(name, slug string, priority int, clock Clock) error {
	name = strings.TrimSpace(name)
	slug = normalizeSlug(slug)

	if err := validateFields(name, slug, priority); err != nil {
		return err
	}

	c.name = name
	c.slug = slug
	c.priority = priority
	c.updatedAt = clock()
	return nil
}

func (c *Category) Rename(name string, clock Clock) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrCategoryNameEmpty
	}
	if len(name) > MaxCategoryNameLength {
		return ErrCategoryNameTooLong
	}

	c.name = name
	c.updatedAt = clock()
	return nil
}

func (c *Category) ChangeSlug(slug string, clock Clock) error {
	slug = normalizeSlug(slug)
	if !slugRegex.MatchString(slug) {
		return ErrCategorySlugInvalid
	}

	c.slug = slug
	c.updatedAt = clock()
	return nil
}

func (c *Category) Reprioritize(priority int, clock Clock) error {
	if priority < 0 {
		return ErrCategoryPriorityInvalid
	}

	c.priority = priority
	c.updatedAt = clock()
	return nil
}

func (c *Category) Activate(clock Clock) {
	if !c.isActive {
		c.isActive = true
		c.updatedAt = clock()
	}
}

func (c *Category) Deactivate(clock Clock) {
	if c.isActive {
		c.isActive = false
		c.updatedAt = clock()
	}
}

func normalizeSlug(slug string) string {
	return strings.ToLower(strings.TrimSpace(slug))
}

func validateFields(name, slug string, priority int) error {
	if name == "" {
		return ErrCategoryNameEmpty
	}
	if len(name) > MaxCategoryNameLength {
		return ErrCategoryNameTooLong
	}
	if !slugRegex.MatchString(slug) {
		return ErrCategorySlugInvalid
	}
	if priority < 0 {
		return ErrCategoryPriorityInvalid
	}
	return nil
}
