package service_test

import (
	"testing"

	customermock "hexagonal-modular-sidecar/internal/customer/repository/mock"
	purchasemock "hexagonal-modular-sidecar/internal/purchase/repository/mock"
	rewardmock "hexagonal-modular-sidecar/internal/reward/repository/mock"
	customersvc "hexagonal-modular-sidecar/internal/customer/service"
)

// testMocks bundles the three repository mocks the customer service depends on.
type testMocks struct {
	Customers *customermock.Repository
	Purchases *purchasemock.Repository
	Rewards   *rewardmock.Repository
}

// newService wires a service backed by fresh mocks. Each mock auto-asserts
// its expectations when the test ends (see the t.Cleanup in NewRepository).
func newService(t *testing.T) (*customersvc.Service, *testMocks) {
	mocks := &testMocks{
		Customers: customermock.NewRepository(t),
		Purchases: purchasemock.NewRepository(t),
		Rewards:   rewardmock.NewRepository(t),
	}
	return customersvc.New(mocks.Customers, mocks.Purchases, mocks.Rewards), mocks
}
