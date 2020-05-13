package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"time"
)

const (
	UpdateTip_Nothing = iota
	UpdateTip_ToUpdate
	UpdateTip_Updating
	UpdateTip_Updated
)

var UpdateTip2DivID = []string{
	UpdateTip_Nothing:  "",
	UpdateTip_ToUpdate: "to-update",
	UpdateTip_Updating: "updating",
	UpdateTip_Updated:  "updated",
}

func (ds *docServer) confirmUpdateTip() {
	if ds.updateTip == UpdateTip_Updating {
		return
	}

	d := time.Now().Sub(ds.roughBuildTime)
	needCheckUpdate := d > time.Hour*24*30 || true
	if needCheckUpdate {
		ds.updateTip = UpdateTip_ToUpdate
		ds.newerVersionInstalled = false
	} else if ds.newerVersionInstalled {
		ds.updateTip = UpdateTip_Updated
	} else {
		ds.updateTip = UpdateTip_Nothing
	}
}

var divShownHidden = map[bool]string{false: " hidden", true: ""}

func (ds *docServer) writeUpdateGoldBlock(page *htmlPage) {
	d := time.Now().Sub(ds.roughBuildTime)
	if true || d < time.Hour*24*30 {
		fmt.Fprintf(page, `
<div id="%s" class="gold-update%s">%s</div>
<div id="%s" class="gold-update hidden">%s</div>
<div id="%s" class="gold-update%s">%s</div>`,
			UpdateTip2DivID[UpdateTip_ToUpdate], divShownHidden[ds.updateTip == UpdateTip_ToUpdate], ds.currentTranslation.Text_UpdateTip("ToUpdate"),
			UpdateTip2DivID[UpdateTip_Updating], ds.currentTranslation.Text_UpdateTip("Updating"),
			UpdateTip2DivID[UpdateTip_Updated], divShownHidden[ds.updateTip == UpdateTip_Updated], ds.currentTranslation.Text_UpdateTip("Updated"),
		)
	}
}

// "/update" API.
// - GET: get current update info.
// - POST: do update
func (ds *docServer) updateAPI(w http.ResponseWriter, r *http.Request) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	ds.confirmUpdateTip()

	if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"updateStatus": "%s"}`, UpdateTip2DivID[ds.updateTip])
		return
	}

	if r.Method == http.MethodPost {
		w.WriteHeader(http.StatusAccepted)
		w.Header().Set("Content-Type", "application/json")
		if ds.updateTip == UpdateTip_ToUpdate {
			ds.updateTip = UpdateTip_Updating
			go ds.updateGold()
		}
		fmt.Fprintf(w, `{"updateStatus": "%s"}`, UpdateTip2DivID[ds.updateTip])
		return
	}

	w.WriteHeader(http.StatusNotFound)
}

func (ds *docServer) onUpdateDone(succeeded bool) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	ds.roughBuildTime = time.Now()
	ds.newerVersionInstalled = succeeded
}

func (ds *docServer) updateGold() {
	if err := func() error {
		dir, err := ioutil.TempDir("", "*")
		if err != nil {
			return err
		}

		log.Println("Run: go get -u go101.org/gold")
		output, err := runShellCommand(time.Minute*2, dir, "go", "get", "-u", "go101.org/gold")
		if len(output) > 0 {
			log.Printf("\n%s", output)
		}
		if err != nil {
			return err
		}

		return nil
	}(); err != nil {
		ds.onUpdateDone(false)
		log.Println("Update Gold error:", err)
	} else {
		ds.onUpdateDone(true)
		log.Println("Update Gold succeeded.")
	}
}

func runShellCommand(timeout time.Duration, wd, cmd string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	command := exec.CommandContext(ctx, cmd, args...)
	command.Dir = wd
	return command.Output()
}