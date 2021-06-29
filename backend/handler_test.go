package backend

import "testing"

func TestSimple(t *testing.T) {
	data := []int{
		0,1,2,3,4,5,6,7,8,9,
		10,11,12,13,14,15,16,17,18,19,
		20,21,22,23,24,25,26,27,28,29,
	}
	target := 19
	left := 0
	right := len(data)
	for left < right {
		mid := (left + right) / 2
		if data[mid] == target {
			left = mid + 1
		}else if data[mid] < target {
			left = mid + 1
		}else if data[mid] > target {
			right = mid
		}
	}
	t.Logf("find target: %d", data[left - 1])
}

func TestHandler_Start(t *testing.T) {

}