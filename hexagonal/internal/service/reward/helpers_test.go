package reward_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"

	mailmock "hexagonal/internal/mail/mock"
	"hexagonal/internal/repository"
	customermock "hexagonal/internal/repository/customer/mock"
	repomock "hexagonal/internal/repository/mock"
	rewardmock "hexagonal/internal/repository/reward/mock"
	rewardsvc "hexagonal/internal/service/reward"
)

type testMocks struct {
	Customers *customermock.Repository
	Rewards   *rewardmock.Repository
	Tx        *repomock.Transactor
	Mailer    *mailmock.Sender
}

func newService(t *testing.T) (*rewardsvc.Service, *testMocks) {
	mocks := &testMocks{
		Customers: customermock.NewRepository(t),
		Rewards:   rewardmock.NewRepository(t),
		Tx:        repomock.NewTransactor(t),
		Mailer:    mailmock.NewSender(t),
	}
	return rewardsvc.New(mocks.Customers, mocks.Rewards, mocks.Tx, mocks.Mailer), mocks
}

func (m *testMocks) expectTxRun(ctx context.Context) {
	m.Tx.EXPECT().RunTx(ctx, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(repository.TxRepos) error) error {
			return fn(repository.TxRepos{
				Customers: m.Customers,
				Rewards:   m.Rewards,
			})
		})
}
