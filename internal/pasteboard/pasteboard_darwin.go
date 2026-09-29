package pasteboard

import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

const appKitPath = "/System/Library/Frameworks/AppKit.framework/AppKit"

var (
	loadOnce sync.Once
	loadErr  error

	classPasteboard objc.Class
	classPool       objc.Class

	selGeneralPasteboard = objc.RegisterName("generalPasteboard")
	selChangeCount       = objc.RegisterName("changeCount")
	selTypes             = objc.RegisterName("types")
	selCount             = objc.RegisterName("count")
	selObjectAtIndex     = objc.RegisterName("objectAtIndex:")
	selUTF8String        = objc.RegisterName("UTF8String")
	selNew               = objc.RegisterName("new")
	selDrain             = objc.RegisterName("drain")
)

func loadAppKit() error {
	loadOnce.Do(func() {
		if _, err := purego.Dlopen(appKitPath, purego.RTLD_NOW|purego.RTLD_GLOBAL); err != nil {
			loadErr = fmt.Errorf("loading AppKit: %w", err)
			return
		}
		classPasteboard = objc.GetClass("NSPasteboard")
		classPool = objc.GetClass("NSAutoreleasePool")
	})
	return loadErr
}

// ChangeCount returns the general pasteboard's change counter.
func ChangeCount() (int, bool) {
	if loadAppKit() != nil {
		return 0, false
	}
	pb := objc.ID(classPasteboard).Send(selGeneralPasteboard)
	return objc.Send[int](pb, selChangeCount), true
}

func clipboardTypes() ([]string, error) {
	if err := loadAppKit(); err != nil {
		return nil, err
	}

	// Autorelease pools are per OS thread; migrating before drain crashes.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	pool := objc.ID(classPool).Send(selNew)
	defer pool.Send(selDrain)

	pb := objc.ID(classPasteboard).Send(selGeneralPasteboard)
	list := pb.Send(selTypes)
	if list == 0 {
		return nil, nil
	}
	n := objc.Send[uint](list, selCount)
	types := make([]string, 0, n)
	for i := range n {
		name := list.Send(selObjectAtIndex, i)
		types = append(types, goString(objc.Send[*byte](name, selUTF8String)))
	}
	return types, nil
}

// goString copies a NUL-terminated C string into a Go string.
func goString(p *byte) string {
	if p == nil {
		return ""
	}
	n := 0
	for *(*byte)(unsafe.Add(unsafe.Pointer(p), n)) != 0 {
		n++
	}
	return string(unsafe.Slice(p, n))
}
