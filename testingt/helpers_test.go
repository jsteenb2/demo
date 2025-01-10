package testingt_test

import (
	"fmt"
	"maps"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHelperShowcase(t *testing.T) {
	t.Run("t.TempDir", func(t *testing.T) {
		// temp directory is cleaned up after the test completes
		// handy indeed
		dir := t.TempDir()
		t.Log("testing dir: ", dir)
	})

	t.Run("t.SetEnv", func(t *testing.T) {
		t.Log("env var before sub test executes: ", os.Getenv("THREEVE"))
		defer func() {
			t.Log("env var before after test executes: ", os.Getenv("THREEVE"))
		}()

		t.Run("SetEnv - envs isolated to a test", func(t *testing.T) {
			// Note: these are not safe to use in with t.Parallel
			t.Setenv("THREEVE", "$TEXAS")

			t.Logf("THREEVE=%s", os.Getenv("THREEVE"))
		})
	})

	t.Run("t.Parallel", func(t *testing.T) {
		t.Run("sequential takes as long as all the tests combined", func(t *testing.T) {
			//t.Parallel() // uncomment this and subsequent tests to showcase running in parallel together

			tests := []string{"first", "second", "third", "fourth"}
			for _, tt := range tests {
				t.Run(tt, func(t *testing.T) {
					t.Log("starting test: ", tt)
					time.Sleep(1 * time.Second)
					t.Log("finishing test: ", tt)
				})
			}
		})

		t.Run("with t.Parallel it runs as long as the slowest running test plus scheduling costs", func(t *testing.T) {
			//t.Parallel()

			// takeaway: yes you can potentially speed up test, but it's not a panacea.
			//			 The scheduling cost is significant.

			tests := []string{"first", "second", "third", "fourth", "fifth", "sixth", "seventh", "eighth", "NINE THOUSAND"}
			for _, tt := range tests {
				t.Run(tt, func(t *testing.T) {
					t.Parallel()

					t.Log("starting test: ", tt)
					time.Sleep(1 * time.Second)
					t.Log("finishing test: ", tt)
				})
			}
		})
	})

	t.Run("t.Helper - your best friend 3 in CI", func(t *testing.T) {
		t.Run("without t.Helper", func(t *testing.T) {
			fnWithoutHelper(t)
		})

		t.Run("with t.Helper()", func(t *testing.T) {
			fnWithHelper(t)
		})

		t.Run("logging from within helper function", func(t *testing.T) {
			fnHelperWithoutLog(t)
			fnHelperLog(t)
		})

		t.Run("heterogenous use of t.Helper", func(t *testing.T) {
			dispatchHelper(t, fnWithoutHelper)
		})

		t.Run("descendant t.Helper calls", func(t *testing.T) {
			dispatchHelper(t, func(t *testing.T) {
				dispatchHelper(t, func(t *testing.T) {
					// Ruh ohhhh... what's going on here?
					dispatchHelper(t, func(t *testing.T) {
						t.Helper() // got to have that t.Helper or it'll show error here

						t.Fatal("I should bubble up error to the top")
					})
				})
			})
		})

		t.Run("descendant t.Helper calls with testify", func(t *testing.T) {
			dispatchHelper(t, func(t *testing.T) {
				t.Helper()

				dispatchHelper(t, func(t *testing.T) {
					t.Helper()

					dispatchHelper(t, func(t *testing.T) {
						t.Helper() // got to have that t.Helper or it'll show error here

						require.FailNow(t, "I should bubble up error to the top")
					})
				})
			})
		})
	})

	t.Run("t.Cleanup", func(t *testing.T) {
		t.Run("t.Cleanup can be viewed as a defer for the execution of a test", func(t *testing.T) {
			// deferred cleanup fns that will run at the end after test is done
			// can add as many t.Cleanup calls as you'd like
			t.Cleanup(func() { t.Log("|-first log") })
			t.Cleanup(func() { t.Log("|--second log") })
			t.Cleanup(func() { t.Log("|---third log") })
			t.Cleanup(func() { t.Log("|----fourth log") })

			fnWithHelper(t)
		})

		t.Run("useful for isolation in all test scenarios", func(t *testing.T) {
			t.Run("without t.Cleanup we pollute", func(t *testing.T) {
				var store demoStatefulStore
				t.Cleanup(func() {
					t.Log(store.String())
				})
				// sad/bad/happy paths... it doesn't matter. They should all be isolated from one another
				// or risk trying to untangle a gigantic hairball!
				store.Add("first")
				store.Add("second")
				store.Add("third")
			})

			t.Run("with t.Cleanup we don't pollute", func(t *testing.T) {
				var store demoStatefulStore
				t.Cleanup(func() {
					t.Log(store.String())
				})
				// sad/bad/happy paths... it doesn't matter. They should all be isolated from one another
				// or risk trying to untangle a gigantic hairball!
				store.Add("first")
				t.Cleanup(func() { store.Rm("first") })
				store.Add("second")
				t.Cleanup(func() { store.Rm("second") })
				store.Add("third")
				t.Cleanup(func() { store.Rm("third") })
			})

			t.Run("encapsulate cleanup into helper function", func(t *testing.T) {
				var store demoStatefulStore
				t.Cleanup(func() {
					t.Log(store.String())
				})
				// sad/bad/happy paths... it doesn't matter. They should all be isolated from one another
				// or risk trying to untangle a gigantic hairball!
				add(t, &store, "first")
				add(t, &store, "second")
				add(t, &store, "third")

				// can get more powerful with your helper functions
				addKeys(t, &store, "fourth", "fifth", "sixth", "seventh")
			})
		})
	})
}

func fnWithoutHelper(t *testing.T) {
	t.Fatal("failing in fnWithoutHelper")
}

func fnWithHelper(t *testing.T) {
	t.Helper() // <--- HERE"S THE MONEY!
	t.Fatal("failing in fnWithHelper")
}

func dispatchHelper(t *testing.T, fns ...func(t *testing.T)) {
	t.Helper()

	for _, fn := range fns {
		fn(t)
	}
}

func fnHelperWithoutLog(t *testing.T) {
	t.Log("logging in fnHelperWithoutLog")
}

func fnHelperLog(t *testing.T) {
	t.Helper()

	t.Log("logging in fnHelperWithLog")
}

type demoStatefulStore struct {
	state map[string]bool
}

func (d *demoStatefulStore) Add(k string) {
	if d.state == nil {
		d.state = make(map[string]bool)
	}
	d.state[k] = true
}

func (d *demoStatefulStore) Rm(k string) {
	delete(d.state, k)
}

func (d *demoStatefulStore) String() string {
	return fmt.Sprint(slices.Collect(maps.Keys(d.state)))
}

func add(t *testing.T, d *demoStatefulStore, k string) {
	d.Add(k)
	t.Cleanup(func() { d.Rm(k) })
}

func addKeys(t *testing.T, d *demoStatefulStore, keys ...string) {
	// can also avoid calling add here and do it in a for loop here.
	// if number of items is huge, may want to add the cleanup to the keys set
	// with something like:
	//
	//	t.Cleanup(func(){
	//		for _, k := range keys {
	//			d.Rm(k)
	//		}
	//	})
	//
	// IMO, its better to stick with reusing add and have a bunch of cleanups
	// when your test data sets are smallish (<1k calls)

	for _, k := range keys {
		add(t, d, k)
	}
}
