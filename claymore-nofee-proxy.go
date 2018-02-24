package main
import(
    "os"
    "fmt"
    "net"
    "strings"
    "encoding/json"
)

var lock_account = ""
var local_port = ""
var remote_address = ""
var remote_port = ""

func main() {
    if len(os.Args) < 5 {
        fmt.Println("Usage: ./proxy [localport] [remotehost] [remoteport] [ETH Wallet]")
        fmt.Println("Example: ./proxy 9999 eth.realpool.org 9999 0x...")
        return
    }
    local_port = os.Args[1]
    remote_address = os.Args[2]
    remote_port = os.Args[3]
    lock_account = os.Args[4]

    fmt.Println("Wallet set:", lock_account)

    l, err := net.Listen("tcp", ":" + local_port)
    if err != nil {
        fmt.Println("Listen error:", err)
        return
    }
    fmt.Println("Start proxy at port:", local_port)
    for {
        c, err := l.Accept()
        if err != nil {
            fmt.Println("Accept error:", err)
            return
        }
        create_proxy(c)
    }
}

func create_proxy(client net.Conn) {
    server, err := net.Dial("tcp", remote_address + ":" + remote_port)
    if err != nil {
        fmt.Println("Connect to pool error:", err)
        return
    }
    fmt.Println("New connection:", client.RemoteAddr())
    go handle_conn(client, server, true)
    go handle_conn(server, client, false)
}

func handle_conn(c1, c2 net.Conn, local2server bool) {
    var map_result map[string] interface {}
    buf := make([]byte, 512)
    defer c2.Close()
    defer c1.Close()
    if local2server {
        defer fmt.Println("Close connection:", c1.RemoteAddr())
    }
    for {
        data_len, err := c1.Read(buf)
        data := buf
        if err != nil {
            fmt.Println("Read error:", err)
            return
        }
        if local2server {
            err = json.Unmarshal(buf[:data_len], &map_result)
            if err != nil {
                fmt.Println("Decode error:", err)
                return
            }
            if v, ok := map_result["method"]; ok && v == "eth_submitLogin" {
                auth_count := map_result["params"].([]interface{})[0].(string)
                fmt.Println("[*]Auth account:", auth_count)
                if auth_count != lock_account {
                    fmt.Println("[-]Devfee detected")
                    fmt.Println("[*]OLD", auth_count)
                    fmt.Println("[*]NEW", lock_account)
                    buf_str := string(buf[:data_len])
                    data = []byte(strings.Replace(buf_str, auth_count, lock_account, 1))
                    data_len = len(data)
                }
            }
        }
        _, err = c2.Write(data[:data_len])
        if err != nil {
            fmt.Println("Write error:", err)
            return
        }
    }
}
