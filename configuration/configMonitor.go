package configuration

import (
	"context"
	"fmt"

	"github.com/ancalabrese/Reload/handlers"
	"github.com/fsnotify/fsnotify"
)

type ConfigMonitor struct {
	ctx               context.Context
	watcher           *fsnotify.Watcher
	configManager     *ConfigCache
	eventChan         chan (fsnotify.Event)
	errChan           chan (error)
	writeEventHandler *handlers.WriteEventHandler
}

var monitor *ConfigMonitor

// GetConfigMonitorInstance returns an singleton instance of ConfigMonitor
// or an error if fsnotify fails to initialize
func GetConfigMonitorInstance(ctx context.Context) (*ConfigMonitor, error) {
	if monitor == nil {
		w, err := fsnotify.NewWatcher()
		if err != nil {
			return nil, fmt.Errorf("error initializing config monitor: %w", err)
		}

		configManager := GetCacheInstance()
		writeEventChannel := make(chan (*handlers.WriteEvent))
		weh := handlers.NewWriteEventHandler(writeEventChannel)

		monitor = &ConfigMonitor{
			ctx:               ctx,
			watcher:           w,
			configManager:     configManager,
			writeEventHandler: weh,
		}

		go monitor.monitorUp()
	}

	return monitor, nil
}

// TrackNew adds the file path to the monitored paths
func (cm *ConfigMonitor) TrackNew(path string, config interface{}) error {
	c, err := NewConfigurationFile(path, config)
	if err != nil {
		return err
	}

	err = cm.watcher.Add(path)
	if err != nil {
		return fmt.Errorf("error adding new resource %s to monitor: %w", path, err)
	}

	cm.configManager.Add(c)

	return nil
}

// Untrack removes a path from the monitored files
func (cm *ConfigMonitor) Untrack(path string) {
	cm.watcher.Remove(path)
	cm.configManager.Remove(path)
}

// Stop monitoring files and close channels
func (cm *ConfigMonitor) Stop() {
	cm.watcher.Close()
	close(cm.eventChan)
	close(cm.errChan)
	monitor = nil
}

// monitorUp starts listening for events.
// When an event is received it is redirected to the correct event handler
func (cm *ConfigMonitor) monitorUp() {
	for {
		select {
		case <-cm.ctx.Done():
			cm.Stop()
			return

		case event := <-cm.watcher.Events:
			if event.Op.Has(fsnotify.Write) {
				writeEvent, _ := handlers.NewWriteEvent(event)
				cm.writeEventHandler.EventChannel <- writeEvent
			}

		case err := <-cm.watcher.Errors:
			//Send any error back to the caller
			cm.errChan <- err
		}
	}
}
