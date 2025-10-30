//go:build integration

package consumer1_test

import (
	"fmt"
	"model"
	"net/url"
	"os"
	"strconv"
	"testing"

	"consumer1"

	"github.com/pact-foundation/pact-go/v2/consumer"
	"github.com/pact-foundation/pact-go/v2/matchers"
	"github.com/stretchr/testify/assert"
)

var Like = matchers.Like
var EachLike = matchers.EachLike
var Term = matchers.Term
var Regex = matchers.Regex
var HexValue = matchers.HexValue
var Identifier = matchers.Identifier
var IPAddress = matchers.IPAddress
var IPv6Address = matchers.IPv6Address
var Timestamp = matchers.Timestamp
var Date = matchers.Date
var Time = matchers.Time
var UUID = matchers.UUID
var ArrayMinLike = matchers.ArrayMinLike

type S = matchers.S
type Map = matchers.MapMatcher

var u *url.URL
var client *consumer1.Client

func TestClientPact_GetUser(t *testing.T) {
	mockProvider, err := consumer.NewV4Pact(consumer.MockHTTPProviderConfig{
		Consumer: os.Getenv("CONSUMER_NAME"),
		Provider: os.Getenv("PROVIDER_NAME"),
		LogDir:   os.Getenv("LOG_DIR"),
		PactDir:  os.Getenv("PACT_DIR"),
	})
	assert.NoError(t, err)

	t.Run("Product exist", func(t *testing.T) {
		id := 10

		err = mockProvider.
			AddInteraction().
			Given("Product exists").
			UponReceiving("A request to get product").
			WithRequestPathMatcher("GET", Regex("/api/v1/products/"+strconv.Itoa(id), "/api/v1/products/[0-9]+")).
			WillRespondWith(200, func(b *consumer.V4ResponseBuilder) {
				b.BodyMatch(model.Product{}).
					Header("Content-Type", Term("application/json", `application\/json`))
			}).
			ExecuteTest(t, func(config consumer.MockServerConfig) error {
				// Act: test our API client behaves correctly

				// Get the Pact mock server URL
				u, _ = url.Parse("http://" + config.Host + ":" + strconv.Itoa(config.Port))

				// Initialise the API client and point it at the Pact mock server
				client = &consumer1.Client{
					BaseURL: u,
				}

				// Execute the API client
				product, err := client.GetProduct(id)

				// Assert
				if product.ID != id {
					return fmt.Errorf("wanted product with ID %d but got %d", id, product.ID)
				}

				return err
			})

		assert.NoError(t, err)

	})

	t.Run("the product does not exist", func(t *testing.T) {
		id := 10

		err = mockProvider.
			AddInteraction().
			Given("Product does not exist").
			UponReceiving("A request to get a product that does not exist").
			WithRequestPathMatcher("GET", Regex("/api/v1/products/"+strconv.Itoa(id), "/api/v1/products/[0-9]+")).
			WillRespondWith(404, func(b *consumer.V4ResponseBuilder) {
				b.Header("Content-Type", Term("application/json", `application\/json`))
			}).
			ExecuteTest(t, func(config consumer.MockServerConfig) error {
				// Get the Pact mock server URL
				u, _ = url.Parse("http://" + config.Host + ":" + strconv.Itoa(config.Port))

				// Initialise the API client and point it at the Pact mock server
				client = &consumer1.Client{
					BaseURL: u,
				}

				// Act: Execute the API client
				_, err := client.GetProduct(id)

				// Assert
				assert.Equal(t, model.ErrNotFound, err)
				return nil
			})
		assert.NoError(t, err)

	})

}
