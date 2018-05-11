package commands

import (
	"encoding/hex"
	"fmt"

	"github.com/decred/politeia/politeiawww/cmd/politeiawwwcli/config"
	"github.com/decred/politeia/util"
)

type NewuserCmd struct {
	Args struct {
		Email    string `positional-arg-name:"email"`
		Password string `positional-arg-name:"password"`
	} `positional-args:"true" optional:"true"`
	Random        bool   `long:"random" optional:"true" description:"Generate a random email/password for the user"`
	Save          bool   `long:"save" optional:"true" description:"Save the user's identity to datadir for future use"`
	Verify        bool   `long:"verify" optional:"true" description:"Verify the user's email address"`
	Paywall       bool   `long:"paywall" optional:"true" description:"Satisfy paywall fee using testnet faucet"`
	OverrideToken string `long:"overridetoken" optional:"true" description:"Override token for the testnet faucet"`
}

func (cmd *NewuserCmd) Execute(args []string) error {
	if !cmd.Random && cmd.Args.Email == "" {
		return fmt.Errorf("You must either provide a email & password or use " +
			"the --random flag")
	}

	// fetch Politeia policy for password requirements
	pr, err := Ctx.Policy()
	if err != nil {
		return err
	}

	// assign email/password
	var email string
	var password string

	if cmd.Random {
		b, err := util.Random(int(pr.MinPasswordLength))
		if err != nil {
			return err
		}

		email = hex.EncodeToString(b) + "@example.com"
		password = hex.EncodeToString(b)
	} else {
		email = cmd.Args.Email
		password = cmd.Args.Password
	}

	// create new user
	token, id, paywallAddress, paywallAmount, err := Ctx.NewUser(email, password)
	if err != nil {
		return err
	}

	// save user identity to HomeDir
	if cmd.Save {
		id.Save(config.UserIdentityFile)
		fmt.Printf("User identity saved to: %v\n", config.UserIdentityFile)
	}

	// verify user's email address
	if cmd.Verify {
		sig := id.SignMessage([]byte(token))
		err = Ctx.VerifyNewUser(email, token, hex.EncodeToString(sig[:]))
		if err != nil {
			return err
		}
	}

	// satisfy paywall fee using testnet faucet
	if cmd.Paywall {
		if paywallAddress == "" && paywallAmount == 0 {
			return fmt.Errorf("unable to pay %v DCR to %v\n", paywallAmount,
				paywallAddress)
		}

		faucetTx, err := util.PayWithTestnetFaucet(config.FaucetURL,
			paywallAddress, paywallAmount, cmd.OverrideToken)
		if err != nil {
			return fmt.Errorf("unable to pay  %v DCR to %v with faucet: %v",
				paywallAmount, paywallAddress, err)
		}

		fmt.Printf("paid %v DCR to %v with faucet tx %v\n", paywallAmount,
			paywallAddress, faucetTx)
	}

	return nil
}
