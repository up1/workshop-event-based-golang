package provider1_test

import (
	"fmt"
	l "log"
	"net"
	"net/http"
	"os"
	"provider1"
	"testing"

	"model"
	"provider1/repository"

	"github.com/pact-foundation/pact-go/v2/log"
	"github.com/pact-foundation/pact-go/v2/models"
	"github.com/pact-foundation/pact-go/v2/provider"
	"github.com/pact-foundation/pact-go/v2/utils"
)

var port, _ = utils.GetFreePort()

// The Provider verification
func TestPactProvider(t *testing.T) {
	log.SetLogLevel("INFO")

	go startInstrumentedProvider()

	verifier := provider.NewVerifier()

	// Verify the Provider - Branch-based Published Pacts for any known consumers
	err := verifier.VerifyProvider(t, provider.VerifyRequest{
		Provider:           "provider1",
		ProviderBaseURL:    fmt.Sprintf("http://127.0.0.1:%d", port),
		ProviderBranch:     os.Getenv("VERSION_BRANCH"),
		FailIfNoPactsFound: false,
		// Use this if you want to test without the Pact Broker
		// PactFiles:                   []string{filepath.FromSlash(fmt.Sprintf("%s/GoAdminService-GoUserService.json", os.Getenv("PACT_DIR")))},
		BrokerURL:                  fmt.Sprintf("%s://%s", os.Getenv("PACT_BROKER_PROTO"), os.Getenv("PACT_BROKER_URL")),
		BrokerUsername:             os.Getenv("PACT_BROKER_USERNAME"),
		BrokerPassword:             os.Getenv("PACT_BROKER_PASSWORD"),
		PublishVerificationResults: true,
		ProviderVersion:            os.Getenv("VERSION_COMMIT"),
		StateHandlers:              stateHandlers,
		BeforeEach: func() error {
			provider1.GproductRepository = productExists
			return nil
		},
	})

	if err != nil {
		t.Log(err)
	}
}

var stateHandlers = models.StateHandlers{
	"Product exists": func(setup bool, s models.ProviderState) (models.ProviderStateResponse, error) {
		provider1.GproductRepository = productExists
		return models.ProviderStateResponse{}, nil
	},
	"Product does not exist": func(setup bool, s models.ProviderState) (models.ProviderStateResponse, error) {
		provider1.GproductRepository = productDoesNotExist
		return models.ProviderStateResponse{}, nil
	},
}

// Starts the provider API with hooks for provider states.
// This essentially mirrors the main.go file, with extra routes added.
func startInstrumentedProvider() {
	mux := provider1.GetHTTPHandler()

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		l.Fatal(err)
	}
	defer ln.Close()

	l.Printf("API starting: port %d (%s)", port, ln.Addr())
	l.Printf("API terminating: %v", http.Serve(ln, mux))

}

// Provider States data sets
var productExists = &repository.ProductRepository{
	Products: map[string]*model.Product{
		"product10": {
			ID:          10,
			ProductName: "Product 10",
			Price:       100,
			Stock:       10,
		},
	},
}

var productDoesNotExist = &repository.ProductRepository{
	Products: map[string]*model.Product{},
}
