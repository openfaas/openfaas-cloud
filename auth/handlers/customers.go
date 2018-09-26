package handlers

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

// customerCacheExpiry matches the CDN value of GitHub for "RAW" files
const customerCacheExpiry = time.Minute * 5

// Customers checks whether users are customers of OpenFaaS Cloud
type Customers struct {
	Usernames *map[string]string
	Sync      *sync.Mutex
	Expires   time.Time
}

// NewCustomers creates a Customers instance
func NewCustomers() *Customers {
	return &Customers{
		Sync:    &sync.Mutex{},
		Expires: time.Now().Add(time.Minute * -1),
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

	customersURL := os.Getenv("customers_url")
	if len(customersURL) == 0 {
		customersURL = "https://raw.githubusercontent.com/openfaas/openfaas-cloud/master/CUSTOMERS"
	}

	log.Printf("Fetching customers from %s", customersURL)
	customers, getErr := getCustomers(customersURL)
	if getErr != nil {
		log.Printf("unable to fetch customers from %s, error: %s", customersURL, getErr.Error())
		return getErr
	}

	usernames := map[string]string{}
	for _, customer := range customers {
		usernames[customer] = "true"
	}

	c.Sync.Lock()
	defer c.Sync.Unlock()

	c.Usernames = &usernames
	c.Expires = time.Now().Add(customerCacheExpiry)

	return nil
}

// getCustomers reads a list of customers separated by new lines
// who are valid users of OpenFaaS cloud
func getCustomers(customerURL string) ([]string, error) {
	customers := []string{}

	if len(customerURL) == 0 {
		return nil, fmt.Errorf("customerURL was nil")
	}

	c := http.Client{}

	httpReq, _ := http.NewRequest(http.MethodGet, customerURL, nil)
	res, reqErr := c.Do(httpReq)

	if reqErr != nil {
		return customers, reqErr
	}

	if res.Body != nil {
		defer res.Body.Close()

		pageBody, _ := ioutil.ReadAll(res.Body)
		customers = strings.Split(string(pageBody), "\n")

		for i, c := range customers {
			customers[i] = strings.ToLower(c)
		}

	}

	return customers, nil
}
