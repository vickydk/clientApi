package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	pb "self/clientAPI/referential"
	"sync"
	"time"
)

type RpcConnection struct {
	sync.Mutex
	rpcClient *grpc.ClientConn
}

var rpcConnection RpcConnection
var gpRouter *mux.Router

type logResp struct {
	Content struct {
		Token    string `json:"token,omitempty"`
		Expires  string `json:"expires,omitempty"`
		RefToken string `json:"refresh_token,omitempty"`
	} `json:"content"`
	Message string `json:"Message"`
	Code    int    `json:"code"`
}

func RpcConnect() (err error) {
	opts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffLinear(100 * time.Millisecond)),
		//grpc_retry.WithCodes(codes.NotFound, codes.Aborted, codes.Unavailable),
		grpc_retry.WithCodes(codes.ResourceExhausted, codes.Unavailable),
		grpc_retry.WithMax(4),
	}
	rpcConnection.rpcClient, err = grpc.Dial("localhost:50052", grpc.WithInsecure(),
		grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(
			grpc_prometheus.StreamClientInterceptor,
			grpc_retry.StreamClientInterceptor(opts...),
			)),
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
			grpc_prometheus.UnaryClientInterceptor,
			grpc_retry.UnaryClientInterceptor(opts...),
			)))
	return err
}

func getReferential(c pb.ReferentialServiceClient, ctx context.Context, tblName string) *pb.ReferentialListResp{
	r, err := c.List(ctx, &pb.ReferentialReq{TblName: tblName})

	if err != nil {
		fmt.Println("could not greet: %v", err)
	}
	return r
}

func main() {
	err := RpcConnect()
	if err != nil {
		log.Fatal("Unable to initialize connection to RPC")
	}

	gpRouter = mux.NewRouter()

	router := mux.NewRouter()
	router.HandleFunc("/", testGrpc)
	router.HandleFunc("/login", login)
	log.Fatal(http.ListenAndServe(":8099", router))
}

func testGrpc(w http.ResponseWriter, r *http.Request) {
	for j:=1; j<2; j ++ {
		go loopGoRutine()
	}
	fmt.Fprintf(w, "DONE")
}

func loopGoRutine(){
	/*var tblName string
	for i := 1; i < 10; i++ {
		if i == 1 {
			tblName = "order_header"
		} else if i == 2 {
			tblName = "order_item"
		} else if i == 3 {
			tblName = "order_payment"
		} else if i == 4 {
			tblName = "master_payment"
		} else if i == 5 {
			tblName = "order_return"
		} else if i == 6 {
			tblName = "order_return_log"
		} else if i == 7 {
			tblName = "order_cancel"
		} else if i == 8 {
			tblName = "order_cancel_log"
		} else if i == 9 {
			tblName = "customer"
		}
		go sendGrpc(tblName)
	}*/
	sendGrpc("order_header")
}

func sendGrpc(tblName string){
	urlpath := "http://127.0.0.1:8099/login?tblname="+tblName
	fmt.Println(time.Now()," - URL: ",urlpath)
	resp, err:= http.Get(urlpath)
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(time.Now()," - URL: ",urlpath, " |resp sendGrpc: ", string(body))
}

func homeLink(w http.ResponseWriter, r *http.Request) {
	resp, err:= http.Get("http://127.0.0.1:8099/login")

	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println("body: ", string(body))

	fmt.Fprintf(w, string(body))
}

func login(w http.ResponseWriter, r *http.Request) {
	tblName, _ := r.URL.Query()["tblname"]
	ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
	c:= pb.NewReferentialServiceClient(rpcConnection.rpcClient)
	res:= getReferential(c, ctx, string(tblName[0]))
	out, err := json.Marshal(res)
	if err != nil {
		panic (err)
	}
	fmt.Fprintf(w, string(out))
}

func postLogin() string {
	formData := url.Values{
		"username": {"admin"},
		"password": {"admin"},
	}

	resp, err := http.PostForm("http://35.240.132.132:8080/login", formData)

	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		//handle read response error
	}

	messages := logResp{} // Slice of Message instances
	json.Unmarshal(body, &messages)
	return messages.Content.Token
}
