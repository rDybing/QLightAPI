package api

type modeT int

const (
	clientSP modeT = iota
	clientComp
	clientIOT
	ctrlLite
	ctrlPro
	noneSP
	noneComp
)

type configT struct {
	FullChain string
	PrivKey   string
	Local     bool
	AuthID    string
	AuthKey   string
	serverIP  string
}

type welcomeT struct {
	Msg []string
}

type appInfoT struct {
	ID            string
	Name          string
	WH            string
	Aspect        string
	LastPublicIP  string
	LastPrivateIP string
	OS            string
	Model         string
	Logins        int
	FirstLogin    string
	LastLogin     string
	LastUpdate    string
	LastMode      modeT
}

type ipT struct {
	private string
	public  string
}

type loggerT struct {
	Date     string
	Function string
	AppID    string
	Status   string
}

type appListT map[string]appInfoT
