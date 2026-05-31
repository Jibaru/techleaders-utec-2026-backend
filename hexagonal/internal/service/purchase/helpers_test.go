package purchase_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"hexagonal/internal/repository"
	customermock "hexagonal/internal/repository/customer/mock"
	repomock "hexagonal/internal/repository/mock"
	purchasemock "hexagonal/internal/repository/purchase/mock"
	purchasesvc "hexagonal/internal/service/purchase"
)

func mustTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

type testMocks struct {
	Customers *customermock.Repository
	Purchases *purchasemock.Repository
	Tx        *repomock.Transactor
}

func newService(t *testing.T) (*purchasesvc.Service, *testMocks) {
	mocks := &testMocks{
		Customers: customermock.NewRepository(t),
		Purchases: purchasemock.NewRepository(t),
		Tx:        repomock.NewTransactor(t),
	}
	return purchasesvc.New(mocks.Customers, mocks.Purchases, mocks.Tx), mocks
}

// expectTxRun mocks Transactor.RunTx so the service's tx callback actually
// runs against the same repo mocks the test sets expectations on.
func (m *testMocks) expectTxRun(ctx context.Context) {
	m.Tx.EXPECT().RunTx(ctx, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(repository.TxRepos) error) error {
			return fn(repository.TxRepos{
				Customers: m.Customers,
				Purchases: m.Purchases,
			})
		})
}
