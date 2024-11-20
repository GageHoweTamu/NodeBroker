
// create instance of ethclient and assign to cl
cl, err := ethclient.Dial("http://localhost:8545") // is doing this by localhost secure?
if err != nil {
	fmt.Printf("Error dialing ethclient\n")
	panic(err)
}

chainid, err := cl.ChainID(context.Background())
if err != nil {
	panic(err)
}
fmt.Printf("Chain id: %d\n", chainid)
_ = cl
