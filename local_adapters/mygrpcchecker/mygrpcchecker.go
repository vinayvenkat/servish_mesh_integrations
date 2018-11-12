// Copyright 2018 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// nolint:lll
// Generates the mygrpcadapter adapter's resource yaml. It contains the adapter's configuration, name, supported template
// names (metric in this case), and whether it is session or no-session based.
//go:generate $GOPATH/src/istio.io/istio/bin/mixer_codegen.sh -a mixer/adapter/mygrpcchecker/config/config.proto -x "-s=false -n mygrpcchecker -t listentry"

package mygrpcchecker

import (
    "context"
    "fmt"
    "net"

    "google.golang.org/grpc"
    "bytes"

    "istio.io/api/mixer/adapter/model/v1beta1"
    policy "istio.io/api/policy/v1beta1"
    "istio.io/istio/mixer/adapter/mygrpcchecker/config"
    "istio.io/istio/mixer/template/listentry"
    "istio.io/istio/pkg/log"
    "os"
)

type (
    // Server is basic server interface
    Server interface {
        Addr() string
        Close() error
        Run(shutdown chan error)
    }

    // MyGrpcAdapter supports metric template.
    MyGrpcAdapter struct {
        listener net.Listener
        server   *grpc.Server
    }
)

var _ listentry.HandleListEntryServiceServer = &MyGrpcAdapter{}

// HandleLogEntry records log entries
func (s *MyGrpcAdapter) HandleListEntry(ctx context.Context, in *listentry.HandleListEntryRequest) (*v1beta1.CheckResult, error) {

	var b bytes.Buffer
	log.Infof("This is the HandleListEntry function being invoked")
    cfg := &config.Params{}

    if in.AdapterConfig != nil {
        if err := cfg.Unmarshal(in.AdapterConfig.Value); err != nil {
            log.Errorf("error unmarshalling adapter config: %v", err)
            return nil, err
        }
    }

	b.WriteString(fmt.Sprintf("HandleListEntry invoked with:\n  Adapter config: %s\n  Instances: %s\n",
        cfg.String(), in.Instance))

	if cfg.FilePath == "" {
        fmt.Println(b.String())
    } else {
			var ret int
			if _, err := os.Stat(cfg.FilePath); os.IsNotExist(err) {
				fmt.Println("File does not exist. Create it...")
				ret = 1
			}

			if ret == 1 {
				_, err := os.OpenFile(cfg.FilePath, os.O_RDONLY|os.O_CREATE, 0666)
				if err != nil {
					log.Errorf("error creating file: %v", err)
				}

			}

			f, err:= os.OpenFile(cfg.FilePath, os.O_APPEND|os.O_WRONLY, 0666)
			if err != nil {
				log.Errorf("error opening file for append: %v", err)
				return nil, err
			}
		defer f.Close()

        log.Infof("writing instances to file %s", f.Name())
        if _, err = f.Write(b.Bytes()); err != nil {
            log.Errorf("error writing to file: %v", err)
        }
	}
    return &v1beta1.CheckResult{}, nil
}

func decodeDimensions(in map[string]*policy.Value) map[string]interface{} {
    out := make(map[string]interface{}, len(in))
    for k, v := range in {
        out[k] = decodeValue(v.GetValue())
    }
    return out
}

func decodeValue(in interface{}) interface{} {
    switch t := in.(type) {
    case *policy.Value_StringValue:
        return t.StringValue
    case *policy.Value_Int64Value:
        return t.Int64Value
    case *policy.Value_DoubleValue:
        return t.DoubleValue
    case *policy.Value_IpAddressValue:
        ipV := t.IpAddressValue.Value
        ipAddress := net.IP(ipV)
        str := ipAddress.String()
        return str
    case *policy.Value_DurationValue:
        return t.DurationValue.Value.String()
    default:
        return fmt.Sprintf("%v", in)
    }
}

// Addr returns the listening address of the server
func (s *MyGrpcAdapter) Addr() string {
    return s.listener.Addr().String()
}

// Run starts the server run
func (s *MyGrpcAdapter) Run(shutdown chan error) {
    shutdown <- s.server.Serve(s.listener)
}

// Close gracefully shuts down the server; used for testing
func (s *MyGrpcAdapter) Close() error {
    if s.server != nil {
        s.server.GracefulStop()
    }

    if s.listener != nil {
        _ = s.listener.Close()
    }

    return nil
}

// NewMyGrpcAdapter creates a new IBP adapter that listens at provided port.
func NewMyGrpcAdapter(addr string) (Server, error) {
    if addr == "" {
        addr = "0"
    }
    listener, err := net.Listen("tcp", fmt.Sprintf(":%s", addr))
    if err != nil {
        return nil, fmt.Errorf("unable to listen on socket: %v", err)
    }
    s := &MyGrpcAdapter{
        listener: listener,
    }
    fmt.Printf("listening on \"%v\"\n", s.Addr())
    s.server = grpc.NewServer()
    listentry.RegisterHandleListEntryServiceServer(s.server, s)
    return s, nil
}
