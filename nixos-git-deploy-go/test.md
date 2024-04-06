	// signal.Ignore(syscall.SIGTERM)

	// //Wait for the command to finish
	// err = cmd.Wait()
	// if err != nil {
	// 	fmt.Println("Command failed:", err)
	// }
	// var uid = uint32(syscall.Getuid())
	// var gid = uint32(syscall.Getgid())
	// var cred = &syscall.Credential{Uid: uid, Gid: gid}
	// var sysproc = &syscall.SysProcAttr{Credential: cred, Noctty: true}
	// var attr = os.ProcAttr{
	// 	Dir:   ".",
	// 	Env:   os.Environ(),
	// 	Files: []*os.File{os.Stdin, nil, nil},
	// 	Sys:   sysproc,
	// }
	// process, err := os.StartProcess("nix-shell -p --run sleep", []string{"1"}, &attr)
	// //process, err := os.StartProcess(os.Args[0], []string{os.Args[0], "myFunction"}, &attr)
	// if err == nil {
	// 	fmt.Println("Child Process ID:", process.Pid)
	// 	err = process.Release()
	// 	if err != nil {
	// 		fmt.Println(err.Error())
	// 	}
	// } else {
	// 	fmt.Println(err.Error())
	// }

	// // Ignore SIGTERM signal in the parent process
	// signal.Ignore(syscall.SIGTERM)

	// Start a goroutine to run myFunction in a loop for 100 seconds
	// go func() {
	// 	startTime := time.Now()
	// 	for {
	// 		// Run myFunction
	// 		myFunction()

	// 		// Check if 100 seconds have passed
	// 		if time.Since(startTime) >= 100*time.Second {
	// 			break
	// 		}

	// 		// Sleep for 1 second before running myFunction again
	// 		time.Sleep(time.Second)
	// 	}
	// }()