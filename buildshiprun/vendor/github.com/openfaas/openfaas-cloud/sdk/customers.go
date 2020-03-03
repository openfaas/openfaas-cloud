package sdk

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// ValidateCustomers checks environmental
// variable validate_customers if customer
// validation is explicitly disabled
func ValidateCustomers() bool {
	if val, exists := os.LookupEnv("validate_customers"); exists {
		return val != "false" && val != "0"
	}
	return true
}

//ValidateCustomerList validate customer names list
func ValidateCustomerList(customers []string) bool {
	for i, customerName := range customers {
		for j, cn := range customers {

			if i != j {
				if strings.HasPrefix(cn, customerName+"-") {
					return false
				}
			}
		}
	}

	return true
}

// customerCacheExpiry matches the CDN value of GitHub for "RAW" files
const customerCacheExpiry = time.Minute * 5

// Customers checks whether users are customers of OpenFaaS Cloud
type Customers struct {
	Usernames *map[string]string
	Sync      *sync.Mutex
	Expires   time.Time

	CustomersURL  string
	CustomersPath string
}

// NewCustomers creates a Customers struct to be used to query
// valid users.
func NewCustomers(customersPath, customersURL string) *Customers {
	return &Customers{
		Sync:          &sync.Mutex{},
		Expires:       time.Now().Add(time.Minute * -1),
		CustomersPath: customersPath,
		CustomersURL:  customersURL,
	}
}

// Get returns whether a customer is found
func (c *Customers) Get(login string) (bool, error) {
	found := false

	log.Printf("CUSTOMERS cache expires in: %fs", c.Expires.Sub(time.Now()).Seconds())
	if c.Expires.Before(time.Now()) {
		c.Fetch()
	}

	c.Sync.Lock()
	defer c.Sync.Unlock()

	lookup := *c.Usernames

	if _, ok := lookup[strings.ToLower(login)]; ok {
		found = true
	}

	return found, nil
}

// Fetch refreshes cache of customers which is valid for
// `customerCacheExpiry` duration.
func (c *Customers) Fetch() error {
	usernames := map[string]string{}

	if len(c.CustomersPath) > 0 {
		if out, err := ioutil.ReadFile(c.CustomersPath); err == nil {
			values := string(out)

			for _, customer := range strings.Split(values, "\n") {
				if formatted := formatUsername(customer); len(formatted) > 0 {
					usernames[formatted] = "true"
				}
			}
		}
	} else {
		customersURL := os.Getenv("customers_url")
		if len(customersURL) == 0 {
			customersURL = "https://raw.githubusercontent.com/openfaas/openfaas-cloud/master/CUSTOMERS"
		}

		log.Printf("Fetching customers from %s", customersURL)
		customers, getErr := fetchCustomers(customersURL)
		if getErr != nil {
			log.Printf("unable to fetch customers from %s, error: %s", customersURL, getErr.Error())
			return getErr
		}

		for _, customer := range customers {
			usernames[customer] = "true"
		}
	}

	c.Sync.Lock()
	defer c.Sync.Unlock()

	log.Printf("%d customers found", len(usernames))

	c.Usernames = &usernames
	c.Expires = time.Now().Add(customerCacheExpiry)

	return nil
}

// fetchCustomers reads a list of customers separated by new lines
// who are valid users of OpenFaaS cloud
func fetchCustomers(customerURL string) ([]string, error) {
	customers := []string{}

	if len(customerURL) == 0 {
		return nil, fmt.Errorf("customerURL was nil")
	}

	httpReq, _ := http.NewRequest(http.MethodGet, customerURL, nil)
	res, reqErr := http.DefaultClient.Do(httpReq)

	if reqErr != nil {
		return customers, reqErr
	}

	if res.Body != nil {
		defer res.Body.Close()

		pageBody, _ := ioutil.ReadAll(res.Body)

		for _, c := range strings.Split(string(pageBody), "\n") {
			if formatted := formatUsername(c); len(formatted) > 0 {
				customers = append(customers, formatted)
			}
		}
	}

	return customers, nil
}

func formatUsername(input string) string {
	return strings.TrimSpace(strings.ToLower(input))
}
