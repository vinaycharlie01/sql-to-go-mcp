package generator

import (
	"context"
	"fmt"
	"strings"

	"github.com/vinaycharlie01/sql-to-go-mcp/internal/domain/entity"
)

const unitTestTmpl = `package {{.Package}}_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"{{.ModulePath}}/internal/domain/entity"
	"{{.ModulePath}}/internal/domain/repository"
)

// Mock{{.GoName}}Repository is a mock implementation of {{.GoName}}Repository.
type Mock{{.GoName}}Repository struct {
	mock.Mock
}

func (m *Mock{{.GoName}}Repository) Create(ctx context.Context, {{.VarName}} *entity.{{.GoName}}) error {
	args := m.Called(ctx, {{.VarName}})
	return args.Error(0)
}

func (m *Mock{{.GoName}}Repository) GetByID(ctx context.Context, id {{.IDType}}) (*entity.{{.GoName}}, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.{{.GoName}}), args.Error(1)
}

func (m *Mock{{.GoName}}Repository) Update(ctx context.Context, {{.VarName}} *entity.{{.GoName}}) error {
	args := m.Called(ctx, {{.VarName}})
	return args.Error(0)
}

func (m *Mock{{.GoName}}Repository) Delete(ctx context.Context, id {{.IDType}}) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *Mock{{.GoName}}Repository) List(ctx context.Context, limit, offset int) ([]*entity.{{.GoName}}, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.{{.GoName}}), args.Error(1)
}

var _ repository.{{.GoName}}Repository = (*Mock{{.GoName}}Repository)(nil)

func Test{{.GoName}}Repository_Create(t *testing.T) {
	tests := []struct {
		name    string
		input   *entity.{{.GoName}}
		mockErr error
		wantErr bool
	}{
		{
			name:    "success",
			input:   &entity.{{.GoName}}{},
			mockErr: nil,
			wantErr: false,
		},
		{
			name:    "db error",
			input:   &entity.{{.GoName}}{},
			mockErr: fmt.Errorf("connection refused"),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &Mock{{.GoName}}Repository{}
			repo.On("Create", mock.Anything, tc.input).Return(tc.mockErr)

			err := repo.Create(context.Background(), tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			repo.AssertExpectations(t)
		})
	}
}

func Test{{.GoName}}Repository_GetByID(t *testing.T) {
	want := &entity.{{.GoName}}{}
	tests := []struct {
		name    string
		id      {{.IDType}}
		mockRet *entity.{{.GoName}}
		mockErr error
		wantErr bool
	}{
		{
			name:    "found",
			id:      1,
			mockRet: want,
			mockErr: nil,
			wantErr: false,
		},
		{
			name:    "not found",
			id:      999,
			mockRet: nil,
			mockErr: fmt.Errorf("record not found"),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &Mock{{.GoName}}Repository{}
			repo.On("GetByID", mock.Anything, tc.id).Return(tc.mockRet, tc.mockErr)

			got, err := repo.GetByID(context.Background(), tc.id)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.mockRet, got)
			}
			repo.AssertExpectations(t)
		})
	}
}

func Test{{.GoName}}Repository_Update(t *testing.T) {
	input := &entity.{{.GoName}}{}
	tests := []struct {
		name    string
		mockErr error
		wantErr bool
	}{
		{"success", nil, false},
		{"db error", fmt.Errorf("deadlock"), true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &Mock{{.GoName}}Repository{}
			repo.On("Update", mock.Anything, input).Return(tc.mockErr)

			err := repo.Update(context.Background(), input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			repo.AssertExpectations(t)
		})
	}
}

func Test{{.GoName}}Repository_Delete(t *testing.T) {
	tests := []struct {
		name    string
		id      {{.IDType}}
		mockErr error
		wantErr bool
	}{
		{"success", 1, nil, false},
		{"not found", 999, fmt.Errorf("record not found"), true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &Mock{{.GoName}}Repository{}
			repo.On("Delete", mock.Anything, tc.id).Return(tc.mockErr)

			err := repo.Delete(context.Background(), tc.id)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			repo.AssertExpectations(t)
		})
	}
}

func Test{{.GoName}}Repository_List(t *testing.T) {
	want := []*entity.{{.GoName}}{
		{},
	}
	tests := []struct {
		name    string
		limit   int
		offset  int
		mockRet []*entity.{{.GoName}}
		mockErr error
		wantErr bool
	}{
		{"success", 10, 0, want, nil, false},
		{"db error", 10, 0, nil, fmt.Errorf("timeout"), true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &Mock{{.GoName}}Repository{}
			repo.On("List", mock.Anything, tc.limit, tc.offset).Return(tc.mockRet, tc.mockErr)

			got, err := repo.List(context.Background(), tc.limit, tc.offset)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.mockRet, got)
			}
			repo.AssertExpectations(t)
		})
	}
}
`

const integrationTestTmpl = `package {{.Package}}_integration_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"{{.ModulePath}}/internal/adapters/db/repository"
	"{{.ModulePath}}/internal/domain/entity"
)

// setupTestDB connects to a test database.
// In CI, use testcontainers-go; locally, point at a running instance.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=testdb sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err, "open test database")

	require.NoError(t, db.AutoMigrate(&entity.{{.GoName}}{}), "auto migrate {{.TableName}}")
	t.Cleanup(func() {
		db.Exec("TRUNCATE TABLE {{.TableName}} RESTART IDENTITY CASCADE")
	})
	return db
}

func TestIntegration_{{.GoName}}Repository_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	repo := repository.New{{.GoName}}Repository(db, nil)
	ctx := context.Background()

	// Create
	rec := &entity.{{.GoName}}{}
	err := repo.Create(ctx, rec)
	require.NoError(t, err)

	// GetByID
	fetched, err := repo.GetByID(ctx, {{.ZeroID}})
	require.NoError(t, err)
	assert.NotNil(t, fetched)

	// Update
	err = repo.Update(ctx, fetched)
	require.NoError(t, err)

	// List
	list, err := repo.List(ctx, 10, 0)
	require.NoError(t, err)
	assert.NotEmpty(t, list)

	// Delete
	err = repo.Delete(ctx, {{.ZeroID}})
	require.NoError(t, err)

	// Verify deleted
	_, err = repo.GetByID(ctx, {{.ZeroID}})
	assert.Error(t, err, "record should not exist after deletion")

	fmt.Printf("{{.GoName}} CRUD integration test passed\n")
}
`

const mocksTmpl = `package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"{{.ModulePath}}/internal/domain/entity"
)

// {{.GoName}}Repository is a type-safe mock for repository.{{.GoName}}Repository.
type {{.GoName}}Repository struct {
	mock.Mock
}

func (m *{{.GoName}}Repository) Create(ctx context.Context, {{.VarName}} *entity.{{.GoName}}) error {
	return m.Called(ctx, {{.VarName}}).Error(0)
}

func (m *{{.GoName}}Repository) GetByID(ctx context.Context, id {{.IDType}}) (*entity.{{.GoName}}, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.{{.GoName}}), args.Error(1)
}

func (m *{{.GoName}}Repository) Update(ctx context.Context, {{.VarName}} *entity.{{.GoName}}) error {
	return m.Called(ctx, {{.VarName}}).Error(0)
}

func (m *{{.GoName}}Repository) Delete(ctx context.Context, id {{.IDType}}) error {
	return m.Called(ctx, id).Error(0)
}

func (m *{{.GoName}}Repository) List(ctx context.Context, limit, offset int) ([]*entity.{{.GoName}}, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.{{.GoName}}), args.Error(1)
}
`

type testTmplData struct {
	Package    string
	ModulePath string
	GoName     string
	VarName    string
	TableName  string
	IDType     string
	ZeroID     string
}

// TestGenerator generates test files for a repository.
type TestGenerator struct{}

// NewTestGenerator returns a new TestGenerator.
func NewTestGenerator() *TestGenerator { return &TestGenerator{} }

// GenerateUnitTests renders unit test file for the given table.
func (g *TestGenerator) GenerateUnitTests(_ context.Context, table *entity.TableDefinition, opts entity.GeneratorOptions) (entity.GeneratedFile, error) {
	data := buildTestData(table, opts)

	out, err := renderTemplate("unit_tests", unitTestTmpl, data)
	if err != nil {
		return entity.GeneratedFile{}, fmt.Errorf("render unit tests: %w", err)
	}

	return entity.GeneratedFile{
		Path:    fmt.Sprintf("internal/domain/repository/%s_repository_test.go", strings.ToLower(table.Name)),
		Content: out,
	}, nil
}

// GenerateIntegrationTests renders integration test file for the given table.
func (g *TestGenerator) GenerateIntegrationTests(_ context.Context, table *entity.TableDefinition, opts entity.GeneratorOptions) (entity.GeneratedFile, error) {
	data := buildTestData(table, opts)

	out, err := renderTemplate("integration_tests", integrationTestTmpl, data)
	if err != nil {
		return entity.GeneratedFile{}, fmt.Errorf("render integration tests: %w", err)
	}

	return entity.GeneratedFile{
		Path:    fmt.Sprintf("internal/adapters/db/repository/%s_repository_integration_test.go", strings.ToLower(table.Name)),
		Content: out,
	}, nil
}

// GenerateMocks renders a testify mock for the given table's repository.
func (g *TestGenerator) GenerateMocks(_ context.Context, table *entity.TableDefinition, opts entity.GeneratorOptions) (entity.GeneratedFile, error) {
	data := buildTestData(table, opts)

	out, err := renderTemplate("mocks", mocksTmpl, data)
	if err != nil {
		return entity.GeneratedFile{}, fmt.Errorf("render mocks: %w", err)
	}

	return entity.GeneratedFile{
		Path:    fmt.Sprintf("pkg/testutil/mocks/%s_mock.go", strings.ToLower(table.Name)),
		Content: out,
	}, nil
}

func buildTestData(table *entity.TableDefinition, opts entity.GeneratorOptions) testTmplData {
	pkg := opts.Package
	if pkg == "" {
		pkg = "repository"
	}
	module := opts.ModulePath
	if module == "" {
		module = "github.com/example/app"
	}

	goName := toTitle(table.Name)
	varName := strings.ToLower(goName[:1]) + goName[1:]
	idType := pkIDType(table)

	zeroID := "0"
	if idType == "string" {
		zeroID = `""`
	}

	return testTmplData{
		Package:    pkg,
		ModulePath: module,
		GoName:     goName,
		VarName:    varName,
		TableName:  table.Name,
		IDType:     idType,
		ZeroID:     zeroID,
	}
}
