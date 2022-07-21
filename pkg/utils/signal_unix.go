package utils

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

var signalToOsMap = map[Signal]syscall.Signal{
	SignalInt:  syscall.SIGINT,
	SignalTerm: syscall.SIGTERM,
	SignalUsr2: syscall.SIGUSR2,
}

var signalFromOsMap = map[os.Signal]Signal{
	syscall.SIGINT:  SignalInt,
	syscall.SIGTERM: SignalTerm,
	syscall.SIGUSR2: SignalUsr2,
}

func NotifySignal(c chan<- Signal, sig ...Signal) error {
	if c == nil {
		return fmt.Errorf("NotifySignal using nil channel")
	}

	if len(sig) == 0 {
		return fmt.Errorf("NotifySignal must notify at least 1 signal")
	}

	ch := make(chan os.Signal, cap(c))

	sigs := make([]os.Signal, 0, len(sig))
	for _, s := range sig {
		oss, ok := signalToOsMap[s]
		if !ok {
			return fmt.Errorf("NotifySignal unsupported signal %v", s)
		}
		sigs = append(sigs, oss)
	}

	signal.Notify(ch, sigs...)

	go func() {
		for s := range ch {
			c <- signalFromOsMap[s]
		}
	}()

	return nil
}

func RaiseSignal(pid int, sig Signal) error {
	oss, ok := signalToOsMap[sig]

	if !ok {
		return fmt.Errorf("unsupported signal %v", sig)
	}

	return syscall.Kill(pid, oss)
}
