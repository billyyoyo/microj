package coordinate

import (
	"github.com/billyyoyo/microj/logger"
	"io"
	"net/http"
)

func findLeader(localId string, nodes []ConfigNode) bool {
	for _, node := range nodes {
		if node.Id == localId {
			continue
		}
		resp, err := http.Get(node.getServerAddr() + "/admin/leader")
		if err != nil {
			logger.Errorf("ask node %s/admin/leader ", err, node.getServerAddr())
			continue
		}
		if resp.StatusCode == 200 {
			data, _ := io.ReadAll(resp.Body)
			logger.Infof("ask node %s reply %s", node.Id, string(data))
			return true
		}
	}
	return false
}
