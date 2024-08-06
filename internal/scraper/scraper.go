package scraper

import (
	"context"
	"errors"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"strings"
	"sync"
	"time"
)

// Constant  pre-filled HTTPS addresses.
const (
	LoginPageURL        = "https://lspd.central-roleplay.ro/ucp.php?mode=login&redirect=index.php"
	SearchPageURL       = "https://lspd.central-roleplay.ro/memberlist.php?sk=c&sd=a&form=postform&field=username_list&username=%s&email=&search_group_id=0&joined_select=lt&active_select=lt&count_select=eq&joined=&active=&count=&ip=&mode=searchuser"
	DiscordNameSelector = `//form[@id="viewprofile"]/div/div/dl[@class="left-box details profile-details"]/dt[text()="Discord:"]/following-sibling::dd[1]`
)

const (
	factionUser = "username-coloured"
)

// Scraper manages the web scraping process, including login, profile fetching, and role retrieval.
type Scraper struct {
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.Mutex // Ensures thread safety for context and cache operations
	cache    Cache
	password string
}

// New creates a new Scraper instance.
func New(password string) *Scraper {
	return &Scraper{
		mu: sync.Mutex{},
		cache: Cache{
			data: make(map[string]string),
			mu:   sync.Mutex{},
		},
		password: password,
	}
}

// init initializes the web scraping context and performs the login process.
func (s *Scraper) init() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.ctx != nil && s.cancel != nil {
		s.cancel()
	}

	// Define options for the Chrome browser.
	options := []chromedp.ExecAllocatorOption{chromedp.Flag("headless", true),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (HTML, like Gecko) Chrome/99.0.9999.999 Safari/537.36"),
		chromedp.NoSandbox,
	}

	// Create a new Chrome context.
	allocatorCtx, _ := chromedp.NewExecAllocator(context.Background(), options...)
	ctx, cancel := chromedp.NewContext(allocatorCtx, chromedp.WithLogf(func(s string, i ...interface{}) {}))

	s.ctx = ctx
	s.cancel = cancel

	// Navigate to the login page and perform login.
	err := chromedp.Run(ctx, chromedp.Navigate(LoginPageURL),
		chromedp.Sleep(1*time.Second),
		chromedp.WaitVisible("#login", chromedp.ByID),
		chromedp.WaitVisible(`#login > div:nth-child(1) > div > div > fieldset > dl:nth-child(7) > dd > input.button1`, chromedp.ByQuery),
		chromedp.Focus("#username", chromedp.ByID),
		chromedp.SendKeys("#username", "Los Santos Police Department", chromedp.ByID),
		chromedp.Focus("#password", chromedp.ByID),
		chromedp.SendKeys("#password", s.password, chromedp.ByID),
		chromedp.Click(`#login > div:nth-child(1) > div > div > fieldset > dl:nth-child(7) > dd > input.button1`, chromedp.ByQuery),
	)
	if err != nil {
		return err
	}

	return nil
}

func parseDiscordName(username string) (string, error) {
	splitNick := strings.Split(username, "/")
	parsedName := strings.Split(splitNick[0], " ")
	if len(parsedName) < 2 {
		return "", errors.New("can't parse name")
	} else {
		return parsedName[0] + "+" + parsedName[1], nil
	}
}

// getUserProfileURL retrieves the URL of a user's profile based on their username.
func (s *Scraper) getUserProfileURL(username string) (string, error) {
	sepName, err := parseDiscordName(username)
	if err != nil {
		return "", err
	}

	var ret string
	err = chromedp.Run(s.ctx, chromedp.WaitVisible("#username_logged_in"),
		chromedp.Navigate(fmt.Sprintf(SearchPageURL, sepName)),
	)
	if err != nil {
		return "", err
	}

	// Handle if the user is not registered on the forums.
	if strings.Compare(ret, "No members found for this search criterion.") == 0 {
		return "", errors.New("account not found on the forums")
	}

	var ok bool
	err = chromedp.Run(s.ctx,
		chromedp.Evaluate(`document.querySelector(".`+factionUser+`") !== null`, &ok),
	)
	if !ok {
		return "", errors.New("user doesn't have any forum roles")
	}

	hrefSelector := fmt.Sprintf("#memberlist > tbody > tr > td:nth-child(1) > a.%s", factionUser)

	var href string
	err = chromedp.Run(s.ctx, chromedp.AttributeValue(hrefSelector, "href", &href, nil))
	if err != nil {
		return "", errors.New("couldn't get account's URL")
	}

	return href, nil
}

// FetchUserGroups retrieves the user's forum roles and rank based on their username and Discord ID.
func (s *Scraper) FetchUserGroups(username string, discord string) ([]string, string, error) {
	s.mu.Lock()
	if s.ctx == nil || s.cancel == nil {
		s.mu.Unlock()
		if err := s.init(); err != nil {
			return nil, "", err
		}
	} else {
		s.mu.Unlock()
	}

	profileURL, cached := s.cache.Get(username)
	if !cached {
		var err error
		profileURL, err = s.getUserProfileURL(username)
		if err != nil {
			return nil, "", err
		}

		s.cache.Set(username, profileURL)
	}

	var forumDiscord string
	err := chromedp.Run(s.ctx,
		chromedp.Navigate("https://lspd.central-roleplay.ro/"+profileURL[2:]),
		chromedp.Text(DiscordNameSelector, &forumDiscord, chromedp.NodeVisible),
	)
	if err != nil {
		return nil, "", err
	}

	if strings.Compare(forumDiscord, discord) != 0 {
		return nil, "", errors.New("failed to check discord name on the forums")
	}

	var optionsGroup []*cdp.Node
	err = chromedp.Run(s.ctx, chromedp.Nodes("select option", &optionsGroup))
	if err != nil {
		return nil, "", err
	}

	var groupIds []string
	for _, groupNode := range optionsGroup {
		for _, group := range groupNode.Children {
			groupIds = append(groupIds, group.NodeValue)
		}
	}

	var rank string
	err = chromedp.Run(s.ctx,
		chromedp.WaitVisible(`select[name="g"]`),
		chromedp.Evaluate(`document.querySelector('select[name="g"] option[selected="selected"]').textContent`,
			&rank))
	if err != nil {
		return nil, "", err
	}

	return groupIds, rank, nil
}
