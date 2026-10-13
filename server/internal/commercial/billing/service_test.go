package billing

import (
	"testing"

	stripeintegration "github.com/leamout/leamout/server/internal/integrations/payments/stripe"
)

func TestPlanCodeFromStripePrice(t *testing.T) {
	t.Parallel()
	service := &Service{priceIDs: map[string]string{
		"developer": "price_dev", "pro": "price_pro", "scale": "price_scale",
	}}
	for _,tc:=range []struct{name,price,want string;valid bool}{
		{"developer","price_dev","developer",true},
		{"pro","price_pro","pro",true},
		{"scale","price_scale","scale",true},
		{"unknown","price_other","",false},
		{"missing","","",false},
	} {
		t.Run(tc.name,func(t *testing.T){
			var subscription stripeintegration.Subscription
			subscription.Items.Data=[]stripeintegration.SubscriptionItem{{}}
			subscription.Items.Data[0].Price.ID=tc.price
			got,err:=service.planCodeFromPrice(subscription)
			if (err==nil)!=tc.valid || got!=tc.want {t.Fatalf("got %q, err %v; want %q, valid %v",got,err,tc.want,tc.valid)}
		})
	}
	var empty stripeintegration.Subscription
	if _,err:=service.planCodeFromPrice(empty);err==nil {t.Fatal("expected empty subscription rejection")}
	empty.Items.Data=[]stripeintegration.SubscriptionItem{{},{}}
	if _,err:=service.planCodeFromPrice(empty);err==nil {t.Fatal("expected multi-price subscription rejection")}
}

func TestBillingRedirectOrigin(t *testing.T){
	t.Parallel()
	service:=&Service{frontendURL:"https://leamout.com"}
	for _,tc:=range []struct{url string;valid bool}{
		{"https://leamout.com/billing/success",true},
		{"https://LEAMOUT.com/billing/cancel?next=ok",true},
		{"https://evil.example/capture",false},
		{"http://leamout.com/billing",false},
		{"https://leamout.com.evil.example",false},
		{"https://leamout.com@evil.example",false},
		{"//leamout.com",false},
		{"/billing",false},
	}{
		t.Run(tc.url,func(t *testing.T) {
			err:=service.validateRedirects(tc.url)
			if (err==nil)!=tc.valid {t.Fatalf("URL %q: err=%v, valid=%v",tc.url,err,tc.valid)}
		})
	}
}
