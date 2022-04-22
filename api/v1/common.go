package v1

import (
	// "fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Recon struct {
	HasStarted  bool      `json:"has_started,omitempty"`
	HasFinished bool      `json:"has_finished,omitempty"`
	Job         *ReconJob `json:"job,omitempty"`
	LastChecked int64     `json:"last_checked"`
	Message     string    `json:"message,omitempty"`
	Status      bool      `json:"status,omitempty"`
}

func (re *Recon) ShouldCheck() bool {
	if re.HasFinished == false && re.HasStarted == false {
		return true
	}
	return false
}

func (re *Recon) reset() {
	re.HasStarted = false
	re.HasFinished = false
	re.Job = nil
	re.Status = false
	re.Message = ""
}

func (re *Recon) SetStarted() {
	re.HasFinished = false
	re.HasStarted = true
	re.LastChecked = time.Now().Unix()
}

func (re *Recon) SetFinishedWith(status bool, msg string) {
	re.HasStarted = false
	re.HasFinished = true
	re.Status = status
	re.Message = msg
	re.LastChecked = time.Now().Unix()
}

// makes sense only for reconcile requests where a period check of any other resource is required
func (re *Recon) IsRunning() bool {
	if re.HasFinished == false && re.HasStarted == true {
		return true
	}
	return false
}

func (re *Recon) ShouldRetry(coolingTime int) bool {
	t1 := time.Now().Unix()
	t2 := re.LastChecked
	if t1-t2 >= int64(coolingTime) {
		re.reset()
		return true
	}
	return false
}

func (re *Recon) ConditionStatus() metav1.ConditionStatus {
	if re.HasFinished && re.Status {
		return metav1.ConditionTrue
	}
	if re.HasFinished && !re.Status {
		return metav1.ConditionFalse
	}
	return metav1.ConditionUnknown
}

func (re *Recon) Reason() string {
	if re.HasFinished && re.Status {
		return "Success"
	}
	if re.HasFinished && !re.Status {
		return "Failed"
	}
	if !re.HasFinished && re.HasStarted {
		return "Running"
	}
	return "Unknown"
}

type Bool bool

func (b Bool) Status() metav1.ConditionStatus {
	if b {
		return metav1.ConditionTrue
	}
	return metav1.ConditionFalse
}

type Condition struct {
	Type               string
	Status             string // "True", "False", "Unknown"
	ObservedGeneration int64
	Reason             string
	Message            string
}

type Operations struct {
	Apply  string `json:"create"`
	Delete string `json:"delete"`
}

type Output struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}
