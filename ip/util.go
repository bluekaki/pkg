package ip

import (
	"bufio"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/bluekaki/pkg/errors"
)

// MkZone util to make zone
// name: zone name; url: likes http://ipverse.net/ipblocks/data/countries/us.zone
func MkZone(name, url string) (*Zone, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrapf(err, "http get %s err", url)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "read resp body err")
		}
		return nil, errors.Errorf("got resp err, code: %d,  message: %s", resp.StatusCode, string(body))
	}

	zone := &Zone{Name: name}
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		cidr := scanner.Text()
		if strings.HasPrefix(cidr, "#") {
			continue
		}

		zone.CIDR = append(zone.CIDR, cidr)
	}

	if err = scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "scan resp body err")
	}

	if zone.CIDR == nil {
		return nil, errors.New("scan resp body no cidr found")
	}

	return zone, nil
}
