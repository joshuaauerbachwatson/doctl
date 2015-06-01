package domainrecs

import "github.com/digitalocean/godo"

// Retrieve a domain record.
func Retrieve(client *godo.Client, domain string, id int) (*godo.DomainRecord, error) {
	r, _, err := client.Domains.Record(domain, id)
	if err != nil {
		return nil, err
	}

	return r, err
}