package gptest

type Client interface {
	SetToken(token string)
	Prepare() error
	Ask(string) (string, error)
}

type ClientType = string

const (
	ClientGpt35 = "GPT35"
)

func GetClient(clientType ClientType) Client {
	switch clientType {
	case ClientGpt35:
		return NewGpt35Client()
	}
	return nil
}
