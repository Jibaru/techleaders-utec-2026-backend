package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	mailmock "hexagonal-modular-sidecar/internal/shared/mail/mock"
	"hexagonal-modular-sidecar/internal/shared/repository"
	customermock "hexagonal-modular-sidecar/internal/customer/repository/mock"
	repomock "hexagonal-modular-sidecar/internal/shared/repository/mock"
	purchasemock "hexagonal-modular-sidecar/internal/purchase/repository/mock"
	purchasesvc "hexagonal-modular-sidecar/internal/purchase/service"
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
	Mailer    *mailmock.Sender
}

func newService(t *testing.T) (*purchasesvc.Service, *testMocks) {
	mocks := &testMocks{
		Customers: customermock.NewRepository(t),
		Purchases: purchasemock.NewRepository(t),
		Tx:        repomock.NewTransactor(t),
		Mailer:    mailmock.NewSender(t),
	}
	return purchasesvc.New(mocks.Customers, mocks.Purchases, mocks.Tx, mocks.Mailer), mocks
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

// expectAnyMail accepts any Send call. Use in tests where the email content
// is incidental to what's being verified.
func (m *testMocks) expectAnyMail() {
	m.Mailer.EXPECT().
		Send(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).Maybe()
}
