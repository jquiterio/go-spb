/*
 * @file: utils.go
 * @author: Jorge Quitério
 * @copyright (c) 2021 Jorge Quitério
 * @license: MIT
 */

package hub

func TopicInSubscriber(topic string, sub *Subscriber) bool {
	for _, t := range sub.Topics {
		if t == topic {
			return true
		}
	}
	return false
}
