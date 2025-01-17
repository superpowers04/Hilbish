package main

import (
	"fmt"
	"sync"
	"time"

	"hilbish/util"
	
	rt "github.com/arnodel/golua/runtime"
)

var timers *timersModule
var timerMetaKey = rt.StringValue("hshtimer")

type timersModule struct {
	mu *sync.RWMutex
	wg *sync.WaitGroup
	timers map[int]*timer
	latestID int
	running int
}

func newTimersModule() *timersModule {
	return &timersModule{
		timers: make(map[int]*timer),
		latestID: 0,
		mu: &sync.RWMutex{},
		wg: &sync.WaitGroup{},
	}
}

func (th *timersModule) wait() {
	th.wg.Wait()
}

func (th *timersModule) create(typ timerType, dur time.Duration, fun *rt.Closure) *timer {
	th.mu.Lock()
	defer th.mu.Unlock()

	th.latestID++
	t := &timer{
		typ: typ,
		fun: fun,
		dur: dur,
		channel: make(chan struct{}, 1),
		th: th,
		id: th.latestID,
	}
	t.ud = timerUserData(t)

	th.timers[th.latestID] = t
	
	return t
}

func (th *timersModule) get(id int) *timer {
	th.mu.RLock()
	defer th.mu.RUnlock()

	return th.timers[id]
}

// #interface timers
// create(type, time, callback) -> @Timer
// Creates a timer that runs based on the specified `time` in milliseconds.
// The `type` can either be `hilbish.timers.INTERVAL` or `hilbish.timers.TIMEOUT`
// --- @param type number
// --- @param time number
// --- @param callback function
func (th *timersModule) luaCreate(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(3); err != nil {
		return nil, err
	}
	timerTypInt, err := c.IntArg(0)
	if err != nil {
		return nil, err
	}
	ms, err := c.IntArg(1)
	if err != nil {
		return nil, err
	}
	cb, err := c.ClosureArg(2)
	if err != nil {
		return nil, err
	}

	timerTyp := timerType(timerTypInt)
	tmr := th.create(timerTyp, time.Duration(ms) * time.Millisecond, cb)
	return c.PushingNext1(t.Runtime, rt.UserDataValue(tmr.ud)), nil
}

// #interface timers
// get(id) -> @Timer
// Retrieves a timer via its ID.
// --- @param id number
// --- @returns Timer
func (th *timersModule) luaGet(thr *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.Check1Arg(); err != nil {
		return nil, err
	}
	id, err := c.IntArg(0)
	if err != nil {
		return nil, err
	}

	t := th.get(int(id))
	if t != nil {
		return c.PushingNext1(thr.Runtime, rt.UserDataValue(t.ud)), nil
	}

	return c.Next(), nil
}

// #interface timers
// #field INTERVAL Constant for an interval timer type
// #field TIMEOUT Constant for a timeout timer type
// timeout and interval API
/*
If you ever want to run a piece of code on a timed interval, or want to wait
a few seconds, you don't have to rely on timing tricks, as Hilbish has a
timer API to set intervals and timeouts.

These are the simple functions `hilbish.interval` and `hilbish.timeout` (doc
accessible with `doc hilbish`). But if you want slightly more control over
them, there is the `hilbish.timers` interface. It allows you to get
a timer via ID and control them.

## Timer Object
All functions documented with the `Timer` type refer to a Timer object.

An example of usage:
```
local t = hilbish.timers.create(hilbish.timers.TIMEOUT, 5000, function()
	print 'hello!'
end)

t:start()
print(t.running) // true
```
*/
func (th *timersModule) loader(rtm *rt.Runtime) *rt.Table {
	timerMethods := rt.NewTable()
	timerFuncs := map[string]util.LuaExport{
		"start": {timerStart, 1, false},
		"stop": {timerStop, 1, false},
	}
	util.SetExports(rtm, timerMethods, timerFuncs)

	timerMeta := rt.NewTable()
	timerIndex := func(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
		ti, _ := timerArg(c, 0)

		arg := c.Arg(1)
		val := timerMethods.Get(arg)

		if val != rt.NilValue {
			return c.PushingNext1(t.Runtime, val), nil
		}

		keyStr, _ := arg.TryString()

		switch keyStr {
			case "type": val = rt.IntValue(int64(ti.typ))
			case "running": val = rt.BoolValue(ti.running)
			case "duration": val = rt.IntValue(int64(ti.dur / time.Millisecond))
		}

		return c.PushingNext1(t.Runtime, val), nil
	}

	timerMeta.Set(rt.StringValue("__index"), rt.FunctionValue(rt.NewGoFunction(timerIndex, "__index", 2, false)))
	l.SetRegistry(timerMetaKey, rt.TableValue(timerMeta))

	thExports := map[string]util.LuaExport{
		"create": {th.luaCreate, 3, false},
		"get": {th.luaGet, 1, false},
	}

	luaTh := rt.NewTable()
	util.SetExports(rtm, luaTh, thExports)

	util.SetField(rtm, luaTh, "INTERVAL", rt.IntValue(0))
	util.SetField(rtm, luaTh, "TIMEOUT", rt.IntValue(1))

	return luaTh
}

func timerArg(c *rt.GoCont, arg int) (*timer, error) {
	j, ok := valueToTimer(c.Arg(arg))
	if !ok {
		return nil, fmt.Errorf("#%d must be a timer", arg + 1)
	}

	return j, nil
}

func valueToTimer(val rt.Value) (*timer, bool) {
	u, ok := val.TryUserData()
	if !ok {
		return nil, false
	}

	j, ok := u.Value().(*timer)
	return j, ok
}

func timerUserData(j *timer) *rt.UserData {
	timerMeta := l.Registry(timerMetaKey)
	return rt.NewUserData(j, timerMeta.AsTable())
}
