package ssl

//import (
//	"github.com/go-acme/lego/v4/providers/dns"
//	"os"
//	"pmail/utils/errors"
//	"regexp"
//	"strings"
//)
//
//func GetServerParamsList(serverName string) ([]string, error) {
//	var serverParams []string
//
//	infos, err := os.ReadDir("./")
//	if err != nil {
//		return nil, errors.Wrap(err)
//	}
//
//	upperServerName := strings.ToUpper(serverName)
//
//	for _, info := range infos {
//		if strings.HasPrefix(info.Name(), upperServerName) {
//			serverParams = append(serverParams, info.Name())
//		}
//	}
//	if len(serverParams) != 0 {
//		return serverParams, nil
//	}
//
//	_, err = dns.NewDNSChallengeProviderByName(serverName)
//	if err == nil {
//		return nil, errors.New(serverName + " Not Support")
//	}
//	if strings.Contains(err.Error(), "unrecognized DNS provider") {
//		return nil, err
//	}
//
//	re := regexp.MustCompile(`missing: (.+)`)
//	// namesilo: some credentials information are missing: NAMESILO_API_KEY
//	estr := err.Error()
//	name := re.FindStringSubmatch(estr)
//
//	if len(name) == 2 {
//		names := strings.Split(name[1], ",")
//
//		for _, s := range names {
//			serverParams = append(serverParams, s)
//			SetDomainServerParams(s, "empty")
//		}
//
//	}
//	_, err = dns.NewDNSChallengeProviderByName(serverName)
//
//	return serverParams, err
//}
//
//func SetDomainServerParams(name, value string) {
//	key := name
//	err := os.WriteFile(key, []byte(value), 0644)
//	if err != nil {
//		panic(err)
//	}
//	err = os.Setenv(name+"_FILE", key)
//	if err != nil {
//		panic(err)
//	}
//}
