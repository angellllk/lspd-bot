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

var Password string

const (
	LoginPageURL        = "https://lspd.ro-rp.ro/ucp.php?mode=login&redirect=index.php"
	SearchPageURL       = "https://lspd.ro-rp.ro/memberlist.php?sk=c&sd=a&form=postform&field=username_list&username=%s&email=&search_group_id=0&joined_select=lt&active_select=lt&count_select=eq&joined=&active=&count=&ip=&mode=searchuser"
	DiscordNameSelector = `//form[@id="viewprofile"]/div/div/dl[@class="left-box details profile-details"]/dt[text()="Discord:"]/following-sibling::dd[1]`
)

type Scraper struct {
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex // To ensure thread safety
	cache  map[string]string
}

func New() *Scraper {
	return &Scraper{
		ctx:    nil,
		cancel: nil,
		mu:     sync.Mutex{},
		cache:  make(map[string]string),
	}
}

func (s *Scraper) initScraper() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.ctx != nil && s.cancel != nil {
		s.cancel()
	}

	options := []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", true),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.9999.999 Safari/537.36"),
		chromedp.NoSandbox,
	}

	allocatorCtx, _ := chromedp.NewExecAllocator(context.Background(), options...)
	ctx, cancel := chromedp.NewContext(allocatorCtx, chromedp.WithLogf(func(s string, i ...interface{}) {}))

	s.ctx = ctx
	s.cancel = cancel

	err := chromedp.Run(ctx,
		chromedp.Navigate(LoginPageURL),
		chromedp.WaitVisible("#login", chromedp.ByID),
	)
	if err != nil {
		return err
	}

	err = chromedp.Run(ctx,
		chromedp.Sleep(1*time.Second/2),
		chromedp.WaitVisible("#username", chromedp.ByID),
		chromedp.Focus("#username", chromedp.ByID),
		chromedp.SendKeys("#username", "LSPD", chromedp.ByID),
		chromedp.Focus("#password", chromedp.ByID),
		chromedp.SendKeys("#password", Password, chromedp.ByID),
		chromedp.Click(`#login > fieldset > input.button2.specialbutton`, chromedp.ByQuery),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Scraper) getUserProfileURL(username string) (string, error) {
	parsedName := strings.Split(username, " ")
	forQuery := parsedName[0] + "+" + parsedName[1]

	var ret string
	err := chromedp.Run(s.ctx,
		chromedp.WaitVisible("#username_logged_in"),
		chromedp.Navigate(fmt.Sprintf(SearchPageURL, forQuery)),
	)
	if err != nil {
		return "", err
	}

	if strings.Compare(ret, "No members found for this search criterion.") == 0 {
		return "", errors.New("account not found on the forums")
	}

	class1 := "username-coloured"
	class2 := "username"
	var ok1, ok2 bool
	err = chromedp.Run(s.ctx,
		chromedp.Evaluate(`document.querySelector(".`+class1+`") !== null`, &ok1),
		chromedp.Evaluate(`document.querySelector(".`+class2+`") !== null`, &ok2),
	)

	var href string
	var hrefSelector string

	if ok1 {
		hrefSelector = fmt.Sprintf("#memberlist > tbody > tr > td:nth-child(1) > a.%s", class1)
	} else {
		hrefSelector = fmt.Sprintf("#memberlist > tbody > tr > td:nth-child(1) > a.%s", class2)
	}

	err = chromedp.Run(s.ctx, chromedp.AttributeValue(hrefSelector, "href", &href, nil))
	if err != nil {
		return "", errors.New("couldn't get account's URL")
	}

	return href, nil
}

func (s *Scraper) FetchUserGroups(username string, discord string) ([]string, string, error) {
	s.mu.Lock()
	if s.ctx == nil || s.cancel == nil {
		s.mu.Unlock()
		if err := s.initScraper(); err != nil {
			return nil, "", err
		}
	} else {
		s.mu.Unlock()
	}

	s.mu.Lock()
	profileURL, cached := s.cache[username]
	s.mu.Unlock()

	if !cached {
		var err error
		profileURL, err = s.getUserProfileURL(username)
		if err != nil {
			return nil, "", err
		}

		s.mu.Lock()
		s.cache[username] = profileURL
		s.mu.Unlock()
	}

	var forumDiscord string
	err := chromedp.Run(s.ctx,
		chromedp.Navigate("https://lspd.ro-rp.ro/"+profileURL[2:]),
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

	className := "profile-avatar"

	var avatarNode []*cdp.Node
	err = chromedp.Run(s.ctx,
		chromedp.Nodes(fmt.Sprintf(".%s", className), &avatarNode, chromedp.ByQueryAll),
	)
	if err != nil {
		return nil, "", err
	}

	var rank string
	if len(avatarNode) == 0 {
		err = chromedp.Run(s.ctx,
			chromedp.Text(`dl.left-box.details.profile-details > dd:nth-of-type(2)`, &rank, chromedp.ByQuery),
		)
		if err != nil {
			return nil, "", err
		}
	} else {

	}

	if len(rank) == 0 {
		err = chromedp.Run(s.ctx,
			chromedp.Text(`dl.left-box > dd:nth-of-type(1)`, &rank, chromedp.ByQuery),
		)
		if err != nil {
			return nil, "", err
		}
	}

	return groupIds, rank, nil
}
