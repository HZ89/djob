package scheduler

import "fmt"

// This func return a new analyzer by spec
// spec can be:
// - crontab specs, e.g. "* * * * * *"
//   (second) (minute) (hour) (day of month) (month) (day of week)
// - Descriptors, e.g. "@midnight", "@every 1h"
func prepare(spec string) (analyzer, error) {

	//defer func() {
	//	if recovered := recover(); recovered != nil {
	//		err = fmt.Errorf("%v", recovered)
	//	}
	//}()

	if spec[0] == '@' {
		return prepareDescriptors(spec)
	}

	return nil, nil
}

func prepareDescriptors(spec string) (analyzer, error) {

}