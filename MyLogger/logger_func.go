package main

import . "mqant/logger"

func main() {

	Mqlog.GetDefaultLogger()

	Mqlog.Error("error...")
	Mqlog.Info("info...")
	Mqlog.Debug("debug...")

	Mqlog.Close()

}
