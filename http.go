package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"os"
	"os/exec"
	"time"
)

/*
* 1. Create cert.
* 2. Start server.
* 3. Open browser.
 */
func main() {
	// 0. Parse command line arguments.
	slog.Info("[Init] Start")
	browse, bind, port, tls := func() (bool, string, int, bool) {
		noBrowseFlag := flag.Bool("no-browse", false, "NOT run the web browser after the web server is started.")
		bindFlag := flag.String("b", "127.0.0.1", "Bind address.")
		portFlag := flag.Int("p", 3000, "Listen port.")
		noTlsFlag := flag.Bool("no-tls", false, "NOT use https.")
		flag.Parse()
		return !*noBrowseFlag, *bindFlag, *portFlag, !*noTlsFlag
	}()
	schema := "http"
	if tls {
		schema = "https"
	}

	key := "secret/server.key"
	crt := "secret/server.crt"

	slog.Info("[Init] Arguments", "browse", browse, "bind", bind, "port", port, "schema", schema, "key", key, "crt", crt)
	slog.Info("[Init] Completed")

	// 1. Create cert
	if tls {
		if err := CreateKeyPair(key, crt); err != nil {
			slog.Error("", "err", err)
			os.Exit(1)
		}
	}

	// 3. Open browser (after a few seconds).
	if browse {
		slog.Info("[Browser] Enabled to browse")
		go func() {
			delay := 500
			server := fmt.Sprintf("%s://127.0.0.1:%d", schema, port)

			// Wait for the web server is ready.
			slog.Info("[Browser] Started to wait", "msec", delay)
			time.Sleep(time.Duration(delay) * time.Millisecond)
			slog.Info("[Browser] Finished to wait", "msec", delay)

			// Open the web browser.
			slog.Info("[Browser] Open browser", "server", server)
			err := exec.Command("cmd.exe", "/c", "start", server).Start()
			if err != nil {
				slog.Error("[Browser] Failed to browse", "err", err)
				os.Exit(1)
			}
			slog.Info("[Browser] Completed")
		}()
	} else {
		slog.Info("[Browser] Disabled to browse")
	}

	// 2. Start server.
	func() {
		dir := "public"
		listen := fmt.Sprintf("%s:%d", bind, port)

		// Serve public directory and output access log
		slog.Info("[Server] Configured", "public", dir)
		fs := http.FileServer(http.Dir(dir))
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			iat := time.Now()
			xw := &xResponseWriter{w, http.StatusOK}

			// 1.2. Log after processing
			defer func() {
				slog.Info("[Server] Completed",
					"method", r.Method,
					"path", r.URL,
					"remote_addr", r.RemoteAddr,
					"user_agent", r.UserAgent(),
					"status", xw.Status(),
					"duration(msec)", time.Since(iat).Milliseconds())
			}()

			// 1.1. Processing
			fs.ServeHTTP(xw, r)
		})

		slog.Info("[Server] Started", "listen", listen)
		if tls {
			err := http.ListenAndServeTLS(listen, crt, key, nil)
			if err != nil {
				slog.Error("[Server] Failed to run server", "err", err)
				os.Exit(1)
			}
		} else {
			err := http.ListenAndServe(listen, nil)
			if err != nil {
				slog.Error("[Server] Failed to run server", "err", err)
				os.Exit(1)
			}
		}
		slog.Info("[Server] Completed")
	}()
}

type xResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *xResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *xResponseWriter) Status() int {
	return w.status
}

func CreateKeyPair(keyFilename string, crtFilename string) error {
	pub, prv, err := func() (*ecdsa.PublicKey, *ecdsa.PrivateKey, error) {
		prv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, nil, err
		}

		pub := prv.PublicKey

		enc, err := x509.MarshalPKCS8PrivateKey(prv)
		if err != nil {
			return &pub, prv, err
		}

		ofs, err := os.Create(keyFilename)
		if err != nil {
			return &pub, prv, err
		}

		blk := &pem.Block{Type: "PRIVATE KEY", Bytes: enc}
		if err := pem.Encode(ofs, blk); err != nil {
			return &pub, prv, err
		}

		return &pub, prv, nil
	}()

	if err != nil {
		return err
	}

	err = func() error {
		today := time.Now()
		tpl := x509.Certificate{
			// Required
			SerialNumber: big.NewInt(1),
			// Options
			NotBefore: today,
			NotAfter:  today.AddDate(100, 0, 0),
			// CA
			IsCA:                  true,
			KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign, // | x509.KeyUsageKeyEncipherment when RSA
			BasicConstraintsValid: true,
			// ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},        // is used when RSA (keyEncipherment)
		}

		crt, err := x509.CreateCertificate(rand.Reader, &tpl, &tpl, pub, prv)
		if err != nil {
			return err
		}

		ofs, err := os.Create(crtFilename)
		if err != nil {
			return err
		}

		blk := &pem.Block{Type: "CERTIFICATE", Bytes: crt}
		if err := pem.Encode(ofs, blk); err != nil {
			return err
		}

		return nil
	}()

	if err != nil {
		return err
	}

	return nil
}
