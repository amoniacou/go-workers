package workers

import (
	"github.com/customerio/gospec"
	. "github.com/customerio/gospec"
	"github.com/gomodule/redigo/redis"
)

func ScheduledSpec(c gospec.Context) {
	scheduled := newScheduled(Config.RetryKey)

	was := Config.Namespace
	Config.Namespace = "prod:"

	c.Specify("empties retry queues up to the current time", func() {
		conn := Config.Pool.Get()
		defer conn.Close()

		now := nowToSecondsWithNanoPrecision()

		message1, _ := NewMsg("{\"queue\":\"default\",\"foo\":\"bar1\"}")
		message2, _ := NewMsg("{\"queue\":\"myqueue\",\"foo\":\"bar2\"}")
		message3, _ := NewMsg("{\"queue\":\"default\",\"foo\":\"bar3\"}")

		_, err := conn.Do("zadd", "prod:"+Config.RetryKey, now-60.0, message1.ToJson())
		c.Expect(err, Equals, nil)
		_, err = conn.Do("zadd", "prod:"+Config.RetryKey, now-10.0, message2.ToJson())
		c.Expect(err, Equals, nil)
		_, err = conn.Do("zadd", "prod:"+Config.RetryKey, now+60.0, message3.ToJson())
		c.Expect(err, Equals, nil)

		scheduled.poll()

		defaultCount, _ := redis.Int(conn.Do("llen", "prod:queue:default"))
		myqueueCount, _ := redis.Int(conn.Do("llen", "prod:queue:myqueue"))
		pending, _ := redis.Int(conn.Do("zcard", "prod:"+Config.RetryKey))

		c.Expect(defaultCount, Equals, 1)
		c.Expect(myqueueCount, Equals, 1)
		c.Expect(pending, Equals, 1)
	})

	Config.Namespace = was
}
