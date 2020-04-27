package berbix

import (
	"os"
	"testing"
)

const customerUID = "some_cool_customer_uid"

func TestClient(t *testing.T) {
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

	err = client.OverrideTransaction(tokens, &OverrideTransactionOptions{
		ResponsePayload: "us-dl",
		Flags:           []string{
			"id_under_21",
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

	t.Log(resultsA)

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
