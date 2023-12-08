package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/spf13/cobra"
)

// voter info struct from the ABI of eosio contract
type VoterInfo struct {
	Owner                      eos.AccountName   `json:"owner"`
	Proxy                      eos.AccountName   `json:"proxy"`
	Producers                  []eos.AccountName `json:"producers"`
	Staked                     eos.Int64         `json:"staked"`
	UnpaidVoteshare            eos.Float64       `json:"unpaid_voteshare"`
	UnpaidVoteshareLastUpdated eos.TimePoint     `json:"unpaid_voteshare_last_updated"`
	UnpaidVoteshareChangeRate  eos.Float64       `json:"unpaid_voteshare_change_rate"`
	LastClaimTime              eos.TimePoint     `json:"last_claim_time"`
	LastVoteWeight             eos.Float64       `json:"last_vote_weight"`
	ProxiedVoteWeight          eos.Float64       `json:"proxied_vote_weight"`
	IsProxy                    eos.Bool          `json:"is_proxy"`
	Flags1                     uint32            `json:"flags1"`
	Reserved2                  uint32            `json:"reserved2"`
	Reserved3                  any               `json:"reserved3"`
}

// voteproducer action input
type Voteproducer struct {
	Voter     eos.AccountName   `json:"voter"`
	Proxy     eos.AccountName   `json:"proxy"`
	Producers []eos.AccountName `json:"producers"`
}

// claimgbmvote action input
type Claimgbmvote struct {
	Owner eos.AccountName `json:"owner"`
}

// stakeclaim Account config struct
type Account struct {
	Address    eos.AccountName
	Permission eos.PermissionName
	PrivateKey ecc.PrivateKey
	Proxy      eos.AccountName
}

type VoterCacheItem struct {
	UnpaidVoteshareLastUpdated eos.TimePoint
	LastClaimTime              eos.TimePoint
}

// list of WAX endpoints
var endpoints = []string{
	"api-wax-mainnet.wecan.dev",
	"api.wax.alohaeos.com",
	"api.wax.bountyblok.io",
	"api.wax.greeneosio.com",
	"api.waxsweden.org",
	"apiwax.3dkrender.com",
	"wax.blacklusion.io",
	"wax.cryptolions.io",
	"wax.dapplica.io",
	"wax.defibox.xyz",
	"wax.eosphere.io",
	"wax.eosusa.io",
	"wax.eu.eosamsterdam.net",
	"wax.greymass.com",
	"wax.pink.gg",
}

// map to keep the voter info cached, to avoid unneccessary api calls
var voterInfoCache = map[eos.AccountName]VoterCacheItem{}

// parseConfig parses the config.txt file and returns a list of Account structs
func parseConfig(cfgFile string) []Account {
	// get the full path of cfgFile
	fullPath, err := filepath.Abs(cfgFile)
	if err != nil {
		log.Fatalf("Unable to determine full path of %s %v", cfgFile, err)
	}

	// check if the config file exists
	_, err = os.Stat(fullPath)
	if err != nil {
		log.Fatalf("Config file %v does not exist", cfgFile)
	}

	// read the file contents as a string
	data, err := os.ReadFile(fullPath)
	if err != nil {
		log.Fatalf("Unable to read config file %s %v", fullPath, err)
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	if len(lines) == 0 {
		log.Fatalf("%s is empty", fullPath)
	}

	accounts := make([]Account, 0)

	for i, rawline := range lines {
		// trim spaces from each line
		line := strings.TrimSpace(rawline)

		// ignore comments
		if strings.HasPrefix(line, "#") {
			continue
		}

		// ignore empty lines
		if len(line) == 0 {
			continue
		}

		// split the line into parts
		parts := strings.Split(line, ":")
		if len(parts) != 4 {
			log.Fatalf("Unable to parse config line %d: %v", i, line)
		}

		address, permission, key, proxy := parts[0], parts[1], parts[2], parts[3]
		eccKey, err := ecc.NewPrivateKey(key)
		if err != nil {
			log.Fatalf("Invalid private key on line %d: %v", i, err)
		}

		accounts = append(accounts, Account{eos.AccountName(address), eos.PermissionName(permission), *eccKey, eos.AccountName(proxy)})
	}

	return accounts
}

// fetchLastClaim fetches the last claim time and unpaid voteshare from the voters table in eosio contract
func fetchLastClaim(account eos.AccountName) (*eos.TimePoint, *eos.TimePoint, error) {
	// check if account is cached
	if val, ok := voterInfoCache[account]; ok {
		return &val.LastClaimTime, &val.UnpaidVoteshareLastUpdated, nil
	}

	// pick a random endpoint
	endpoint := endpoints[rand.Intn(len(endpoints))]
	api := eos.New(fmt.Sprintf("https://%s", endpoint))

	log.Printf("Fetching voter info for account %v using %v\n", account, endpoint)

	results, err := api.GetTableRows(context.Background(), eos.GetTableRowsRequest{
		Code:       "eosio",
		Scope:      "eosio",
		Table:      "voters",
		LowerBound: string(account),
		UpperBound: string(account),
		Limit:      1,
		JSON:       true,
	})
	if err != nil {
		log.Printf("Error fetching voter info for account: %v using %v\n", account, endpoint)
		return nil, nil, err
	}

	var rows []VoterInfo
	err = results.JSONToStructs(&rows)
	if err != nil {
		log.Printf("Error decoding voter info for account: %v %v\n", account, err)
		return nil, nil, err
	}

	if len(rows) == 0 {
		log.Printf("Account %v has not voted yet\n", account)
		var lct eos.TimePoint = eos.TimePoint(0)
		var lvsu eos.TimePoint = eos.TimePoint(0)
		return &lct, &lvsu, nil
	}

	// save the row in the cache
	voterInfoCache[account] = VoterCacheItem{
		LastClaimTime:              rows[0].LastClaimTime,
		UnpaidVoteshareLastUpdated: rows[0].UnpaidVoteshareLastUpdated,
	}

	return &rows[0].LastClaimTime, &rows[0].UnpaidVoteshareLastUpdated, nil
}

// run runs the script, check if the account is ready to vote & claim or wait until it is
func run(account Account) {
	lastClaimTime, lastVoteshareUpdated, err := fetchLastClaim(account.Address)
	if err != nil {
		return
	}
	var timeDiff time.Duration

	if lastClaimTime != nil {
		// calculate the time difference between the last claim and now
		timeDiff = time.Since(time.UnixMicro(int64(*lastClaimTime)))
	}

	// if it's less than 24 hours, sleep
	if timeDiff < 24*time.Hour {
		// calculate remaining time until next claim
		remaining := time.Until(time.UnixMicro(int64(*lastClaimTime)).Add(24 * time.Hour))

		// pick the shortest duration between `remaining` and 1 hour
		// and make sure it is not negative
		shortest := math.Min(remaining.Seconds(), 3600)
		if shortest < 0 {
			shortest = 0
		}
		// calculate the sleep time
		sleepTime := (time.Duration(shortest) * time.Second)
		// sleep a max of 60 minutes or until the next claim time
		log.Printf("Account %v Sleeping %v", account.Address, sleepTime)
		// add an extra 5 seconds to fix a bug (?)
		// where the time.Sleep wakes up a few ms too soon
		time.Sleep(sleepTime + (5 * time.Second))
		go run(account)
		return
	}

	// construct the actions
	actions := make([]*eos.Action, 0)

	// add the voting action
	actions = append(actions, &eos.Action{
		Account:       "eosio",
		Name:          "voteproducer",
		Authorization: []eos.PermissionLevel{{Actor: account.Address, Permission: account.Permission}},
		ActionData: eos.NewActionData(Voteproducer{
			Voter:     account.Address,
			Producers: []eos.AccountName{},
			Proxy:     account.Proxy,
		}),
	})

	// if the user has voted before, add the claim action
	if *lastVoteshareUpdated > 0 {
		actions = append([]*eos.Action{{
			Account:       "eosio",
			Name:          "claimgbmvote",
			Authorization: []eos.PermissionLevel{{Actor: account.Address, Permission: account.Permission}},
			ActionData: eos.NewActionData(Claimgbmvote{
				Owner: account.Address,
			}),
		}}, actions...)
	}

	log.Printf("Sending transaction for account %v", account.Address)

	// send the transaction
	transact(account, actions)

	// sleep for 30 seconds after the transaction, then re-run again
	// in case the tx failed in the first time
	// or it will just sleep if it had succeeded
	time.Sleep(30 * time.Second)
	run(account)
}

// prepare and submit a transaction to the blockchain
func transact(account Account, actions []*eos.Action) {
	// pick a random endpoint
	endpoint := endpoints[rand.Intn(len(endpoints))]
	api := eos.New(fmt.Sprintf("https://%s", endpoint))

	keyBag := &eos.KeyBag{}
	keyBag.ImportPrivateKey(context.Background(), account.PrivateKey.String())

	api.SetSigner(keyBag)
	api.SetCustomGetRequiredKeys(func(ctx context.Context, tx *eos.Transaction) ([]ecc.PublicKey, error) {
		return keyBag.AvailableKeys(ctx)
	})

	txOpts := &eos.TxOptions{}
	if err := txOpts.FillFromChain(context.Background(), api); err != nil {
		log.Printf("Error filling tx opts: %v", err)
		return
	}

	tx := eos.NewTransaction(actions, txOpts)

	_, packedTx, err := api.SignTransaction(context.Background(), tx, txOpts.ChainID, eos.CompressionNone)
	if err != nil {
		log.Printf("Error signing transaction: %v", err)
		return
	}

	response, err := api.SendTransaction(context.Background(), packedTx)

	if err != nil {
		log.Printf("Error sending transaction for account %v using %v %v", account.Address, endpoint, err)
		return
	}

	// delete the cache for this account
	delete(voterInfoCache, account.Address)

	log.Printf("Transaction success for account %v: %v", account.Address, response.TransactionID)
}

func main() {
	var cfgFile string

	var rootCmd = &cobra.Command{
		Use:   "stakeclaim",
		Short: "A simple utility to claim & refresh vote strength on WAX",
		Long: `A simple utility to claim vote rewards and refresh vote strength daily
                Made by Benjie (https://github.com/benjiewheeler)`,
		Run: func(cmd *cobra.Command, args []string) {
			// log the pid of this process
			log.Printf("Stakeclaim running: %v", os.Getpid())

			// parse the config.txt file
			accounts := parseConfig(cfgFile)

			var wg sync.WaitGroup

			for _, account := range accounts {
				wg.Add(1)
				go run(account)
			}

			wg.Wait()
		},
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config-file", "./config.txt", "config file to read the account keys from")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
