package db

import (
	"reflect"
	"testing"

	"github.com/uniqush/uniqush-push/push"
)

type simpleDeliveryPoint struct {
	DevToken     string `json:"devtoken"`
	ProviderName string `json:"provider,omitempty"`
	push.BasicDeliveryPoint
}

func (self *simpleDeliveryPoint) PushService() string {
	return "gcm"
}

func (self *simpleDeliveryPoint) Provider() string {
	return self.ProviderName
}

func (self *simpleDeliveryPoint) UniqId() string {
	return self.DevToken
}

type nopPusher struct {
}

func (self *nopPusher) Push(req *push.PushRequest, resChan chan<- *push.PushResult) {
	return
}

type simpleProvider struct {
	ApiKey    string `json:"apikey"`
	OtherInfo string `json:"other"`
	push.BasicProvider
}

func (self *simpleProvider) UniqId() string {
	return self.ApiKey
}

func (self *simpleProvider) PushService() string {
	return "gcm"
}

type simplePushService struct {
	push.BasicPushService
	push.UnmarshalFromMapToStructPushService
	nopPusher
}

func (self *simplePushService) EmptyProvider() push.Provider {
	return &simpleProvider{}
}

func (self *simplePushService) EmptyDeliveryPoint() push.DeliveryPoint {
	return &simpleDeliveryPoint{}
}

func (self *simplePushService) Name() string {
	return "gcm"
}

func testAddDelProvider(db PushDatabase, t *testing.T) {
	ps := &simplePushService{}
	ps.This = ps
	push.RegisterPushService(ps)

	p := &simpleProvider{
		ApiKey: "key",
	}
	err := db.AddProvider(p)
	if err != nil {
		t.Fatal(err)
	}
	err = db.DelProvider(p)
	if err != nil {
		t.Fatal(err)
	}
}

func pairsEq(p1, p2 []*ProviderDeliveryPointPair) bool {
	if len(p1) != len(p2) {
		return false
	}
	for _, pair1 := range p1 {
		found := false
		for _, pair2 := range p2 {
			if reflect.DeepEqual(pair1, pair2) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func testAddPairs(db PushDatabase, t *testing.T) {
	ps := &simplePushService{}
	ps.This = ps
	push.RegisterPushService(ps)

	p := &simpleProvider{
		ApiKey: "key",
	}
	p.ServiceName = "service"
	err := db.AddProvider(p)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = db.DelProvider(p)
		if err != nil {
			t.Fatal(err)
		}
	}()

	dp1 := &simpleDeliveryPoint{
		DevToken: "token1",
	}
	dp1.ServiceName = "service"
	dp1.SubscriberName = "sub"
	dp2 := &simpleDeliveryPoint{
		DevToken: "token2",
	}
	dp2.ServiceName = "service"
	dp2.SubscriberName = "sub"

	pairs := make([]*ProviderDeliveryPointPair, 2)
	pairs[0].DeliveryPoint = dp1
	pairs[1].DeliveryPoint = dp2

	newpairs, err := db.AddPairs(pairs...)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		db.DelDeliveryPoint(nil, dp1)
		db.DelDeliveryPoint(nil, dp2)
	}()

	if len(newpairs) != len(pairs) {
		t.Errorf("not same size")
	}

	for _, pair := range newpairs {
		if pair.Provider == nil {
			t.Errorf("provider is nil")
		}
		if !reflect.DeepEqual(pair.Provider, p) {
			t.Errorf("provider is not the same")
		}
		if !reflect.DeepEqual(pair.DeliveryPoint, dp1) &&
			!reflect.DeepEqual(pair.DeliveryPoint, dp2) {
			t.Errorf("unknown delivery point")
		}
	}

	foundpairs, err := db.LoopUpPairs("service", "sub")
	if err != nil {
		t.Fatal(err)
	}
	if !pairsEq(foundpairs, newpairs) {
		t.Fatal("found different pairs")
	}

	// Add again. should not change the database
	newpairs, err = db.AddPairs(pairs...)
	if err != nil {
		t.Fatal(err)
	}
	foundpairs, err = db.LoopUpPairs("service", "sub")
	if err != nil {
		t.Fatal(err)
	}
	if !pairsEq(foundpairs, newpairs) {
		t.Fatal("found different pairs")
	}
}
