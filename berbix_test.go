package berbix

import (
	"os"
	"testing"
)

const customerUID = "some_cool_customer_uid"

func TestCreateTransaction(t *testing.T) {
	secret := os.Getenv("BERBIX_DEMO_TEST_CLIENT_SECRET")
	host := os.Getenv("BERBIX_DEMO_API_HOST")
	templateKey := os.Getenv("BERBIX_DEMO_TEMPLATE_KEY")

	client := NewClient(secret, &ClientOptions{
		Host: host,
	})

	tokens, err := client.CreateTransaction(&CreateTransactionOptions{
		CustomerUID: customerUID,
		TemplateKey: templateKey,
	})
	if err != nil {
		t.Fatal(err)
	}

	assertTransaction(t, client, tokens)
}

func TestCreateHostedTransaction(t *testing.T) {
	secret := os.Getenv("BERBIX_DEMO_TEST_CLIENT_SECRET")
	host := os.Getenv("BERBIX_DEMO_API_HOST")
	templateKey := os.Getenv("BERBIX_DEMO_TEMPLATE_KEY")

	client := NewClient(secret, &ClientOptions{
		Host: host,
	})

	options := &CreateHostedTransactionOptions{
		CreateTransactionOptions: CreateTransactionOptions{
			CustomerUID: customerUID,
			TemplateKey: templateKey,
		},
	}
	resp, err := client.CreateHostedTransaction(options)
	if err != nil {
		t.Fatal(err)
	}

	if resp.HostedURL == "" {
		t.Error("expected hosted url to be returned")
	}

	assertTransaction(t, client, &resp.Tokens)
}

func assertTransaction(t *testing.T, client Client, tokens *Tokens) {
	err := client.OverrideTransaction(tokens, &OverrideTransactionOptions{
		ResponsePayload: "us-dl",
		Flags: []string{
			"id_under_21",
		},
		OverrideFields: map[string]string{
			"given_name": "the_name",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	resultsA, err := client.FetchTransaction(tokens)
	if err != nil {
		t.Fatal(err)
	}

	if resultsA.CustomerUID != customerUID {
		t.Errorf("customer UID did not match expectations")
	}

	if len(resultsA.Flags) != 1 {
		t.Fatal("number of flags did not match expectations")
	}

	if resultsA.Flags[0] != "id_under_21" {
		t.Errorf("expected id_under_21 flag")
	}

	if resultsA.Fields == nil || resultsA.Fields.GivenName == nil || resultsA.Fields.GivenName.Value != "the_name" {
		t.Errorf("expected GivenName to be the_name but was %s", resultsA.Fields.GivenName)
	}

	refreshToken := TokensFromRefresh(tokens.RefreshToken)

	resultsB, err := client.FetchTransaction(refreshToken)
	if err != nil {
		t.Fatal(err)
	}

	if resultsA.CustomerUID != resultsB.CustomerUID {
		t.Errorf("expected matching customer UID")
	}

	if err := client.DeleteTransaction(tokens); err != nil {
		t.Fatal(err)
	}
}

func TestDefaultClient_ValidateSignature(t *testing.T) {
	client := NewClient("", &ClientOptions{})

	wh_secret := os.Getenv("BERBIX_WEBHOOK_SECRET")
	body := "{\"transaction_id\":1234123412341234,\"customer_uid\":\"unique-uid\",\"action\":\"test-action\",\"dashboard_link\":\"https://docs.berbix.com\",\"id\":1234123412341234}\n"
	header := `v0,1614990541,9731afbbb3ebcc534775bffed585265283a8ec48ba39d19f9295a2e367c0daeb`
	err := client.ValidateSignature(wh_secret, body, header)
	if err != nil {
		t.Fatal(err)
	}
}
