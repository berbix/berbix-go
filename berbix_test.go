package berbix

import (
	"io/ioutil"
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

	assertTransaction(t, client, tokens, true)
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

	assertTransaction(t, client, &resp.Tokens, true)
}

func TestCreateAPIOnlyTransaction(t *testing.T) {
	secret := os.Getenv("BERBIX_DEMO_TEST_CLIENT_SECRET")
	host := os.Getenv("BERBIX_DEMO_API_HOST")
	templateKey := os.Getenv("BERBIX_DEMO_API_ONLY_TEMPLATE_KEY")
	// for simplicity, hardcode assumptions
	const idType = "P"
	const idCountry = "CA"
	frontUploadPath := os.Getenv("BERBIX_SAMPLE_CA_PASSPORT_PATH")

	client := NewClient(secret, &ClientOptions{
		Host: host,
	})

	options := &CreateAPIOnlyTransactionOptions{
		CreateTransactionOptions: CreateTransactionOptions{
			CustomerUID: customerUID,
			TemplateKey: templateKey,
		},
		APIOnlyOptions: APIOnlyOptions{
			IDType:    idType,
			IDCountry: idCountry,
		},
	}
	createRes, err := client.CreateAPIOnlyTransaction(options)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("will upload image at %q", frontUploadPath)
	frontRdr, err := os.Open(frontUploadPath)
	if err != nil {
		t.Fatal(err)
	}
	defer frontRdr.Close()
	frontBytes, err := ioutil.ReadAll(frontRdr)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("total number of bytes in image: %d", len(frontBytes))

	opts := &UploadImagesOptions{
		Images: []RawImage{
			{
				Image:   frontBytes,
				Subject: ImageSubjectDocumentFront,
				Format:  ImageFormatJPEG,
			},
		},
	}
	upRes, err := client.UploadImages(&createRes.Tokens, opts)
	if err != nil {
		t.Fatal(err)
	}

	const expectedNextStep = "done"
	if upRes.NextStep != expectedNextStep {
		t.Errorf("expected next step of %q but got %q", expectedNextStep, upRes.NextStep)
	}

	opts = &UploadImagesOptions{
		Images: []RawImage{
			{
				Image:   frontBytes,
				Subject: ImageSubjectDocumentFront,
				Format:  ImageFormatJPEG,
			},
		},
	}
	_, err = client.UploadImages(&createRes.Tokens, opts)
	if _, ok := err.(InvalidStateErr); !ok {
		t.Errorf("expected invalid state error, got %v", err)
	}

	// Can't override because transaction has already been completed, so just make
	// sure we can get the transaction metadata
	assertCustomerUIDFromAPI(t, client, &createRes.Tokens)
}

func TestOverrideAPIOnlyTransaction(t *testing.T) {
	secret := os.Getenv("BERBIX_DEMO_TEST_CLIENT_SECRET")
	host := os.Getenv("BERBIX_DEMO_API_HOST")
	templateKey := os.Getenv("BERBIX_DEMO_API_ONLY_TEMPLATE_KEY")
	// for simplicity, hardcode assumptions
	const idType = "DL"
	const idCountry = "US"

	client := NewClient(secret, &ClientOptions{
		Host: host,
	})

	options := &CreateAPIOnlyTransactionOptions{
		CreateTransactionOptions: CreateTransactionOptions{
			CustomerUID: customerUID,
			TemplateKey: templateKey,
		},
		APIOnlyOptions: APIOnlyOptions{
			IDType:    idType,
			IDCountry: idCountry,
		},
	}
	createRes, err := client.CreateAPIOnlyTransaction(options)
	if err != nil {
		t.Fatal(err)

	}
	frontUploadPath := os.Getenv("BERBIX_SAMPLE_DL_FRONT_PATH")
	t.Logf("will upload image at %q", frontUploadPath)
	frontRdr, err := os.Open(frontUploadPath)
	if err != nil {
		t.Fatal(err)
	}
	defer frontRdr.Close()
	frontBytes, err := ioutil.ReadAll(frontRdr)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("total number of bytes in image: %d", len(frontBytes))
	opts := &UploadImagesOptions{
		Images: []RawImage{
			{
				Image:   frontBytes,
				Subject: ImageSubjectDocumentFront,
				Format:  ImageFormatJPEG,
			},
		},
	}
	upRes, err := client.UploadImages(&createRes.Tokens, opts)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("upRes: %+v", upRes)
	const deleteTransaction = false
	assertTransaction(t, client, &createRes.Tokens, deleteTransaction)
}

func assertTransaction(t *testing.T, client Client, tokens *Tokens, deleteTransaction bool) {
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

	resultsA := assertCustomerUIDFromAPI(t, client, tokens)

	if len(resultsA.Flags) != 1 {
		t.Fatalf("number of flags did not match expectations. Flags were: %v", resultsA.Flags)
	}

	if resultsA.Flags[0] != "id_under_21" {
		t.Errorf("expected id_under_21 flag")
	}

	if resultsA.Fields == nil || resultsA.Fields.GivenName == nil || resultsA.Fields.GivenName.Value != "the_name" {
		t.Errorf("expected GivenName to be the_name but was %s", resultsA.Fields.GivenName.Value)
	}

	refreshToken := TokensFromRefresh(tokens.RefreshToken)

	resultsB, err := client.FetchTransaction(refreshToken)
	if err != nil {
		t.Fatal(err)
	}

	if resultsA.CustomerUID != resultsB.CustomerUID {
		t.Errorf("expected matching customer UID")
	}

	if deleteTransaction {
		if err := client.DeleteTransaction(tokens); err != nil {
			t.Fatal(err)
		}
	}
}

func assertCustomerUIDFromAPI(t *testing.T, client Client, tokens *Tokens) *TransactionMetadata {
	resultsA, err := client.FetchTransaction(tokens)
	if err != nil {
		t.Fatal(err)
	}

	if resultsA.CustomerUID != customerUID {
		t.Errorf("customer UID did not match expectations")
	}
	return resultsA
}

func TestUploadOversizedImageAPIOnly(t *testing.T) {
	secret := os.Getenv("BERBIX_DEMO_TEST_CLIENT_SECRET")
	host := os.Getenv("BERBIX_DEMO_API_HOST")
	templateKey := os.Getenv("BERBIX_DEMO_API_ONLY_TEMPLATE_KEY")

	client := NewClient(secret, &ClientOptions{
		Host: host,
	})

	options := &CreateAPIOnlyTransactionOptions{
		CreateTransactionOptions: CreateTransactionOptions{
			CustomerUID: customerUID,
			TemplateKey: templateKey,
		},
		APIOnlyOptions: APIOnlyOptions{},
	}
	createRes, err := client.CreateAPIOnlyTransaction(options)
	if err != nil {
		t.Fatal(err)
	}

	tooManyBytes := make([]byte, 11*1024*1024*1024)
	opts := &UploadImagesOptions{
		Images: []RawImage{
			{
				Image:   tooManyBytes,
				Subject: ImageSubjectDocumentFront,
				Format:  ImageFormatJPEG,
			},
		},
	}
	_, err = client.UploadImages(&createRes.Tokens, opts)
	if _, ok := err.(PayloadTooLargeErr); !ok {
		t.Errorf("expected to get a PayloadTooLargeErr, but got %v", err)
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
