package onepassword

// Registration: consumers import this package and use NewCLI or NewConnect
// directly. If secret.Open factory gains registration support, add:
//
//	func init() {
//	    secret.Register("onepassword", func(cfg map[string]string) (secret.Store, error) {
//	        if url := cfg["connect_url"]; url != "" {
//	            return NewConnect(url, cfg["token"], cfg["vault"]), nil
//	        }
//	        return NewCLI(cfg["vault"]), nil
//	    })
//	}
