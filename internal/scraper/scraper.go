package scraper

import (
	"context"
	"errors"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"strings"
	"time"
)

var Password string

const LoginPageURL = "https://lspd.ro-rp.ro/ucp.php?mode=login&redirect=index.php"
const SearchPageURL = "https://lspd.ro-rp.ro/memberlist.php?sk=c&sd=a&form=postform&field=username_list&username=%s&email=&search_group_id=0&joined_select=lt&active_select=lt&count_select=eq&joined=&active=&count=&ip=&mode=searchuser"
const DiscordNameSelector = `//form[@id="viewprofile"]/div/div/dl[@class="left-box details profile-details"]/dt[text()="Discord:"]/following-sibling::dd[1]`

func initScraper() (context.Context, context.CancelFunc, error) {
	// Define custom options for the Chrome instance
	options := []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", true), // Set to true for headless mode
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.9999.999 Safari/537.36"), // Set a custom User-Agent
		chromedp.NoSandbox,
	}

	// Create a new ExecAllocator with the specified options
	allocatorCtx, _ := chromedp.NewExecAllocator(context.Background(), options...)

	// Create a new chromedp context using the allocator context
	ctx, cancel := chromedp.NewContext(allocatorCtx)

	err := chromedp.Run(ctx,
		chromedp.Navigate(LoginPageURL),
		chromedp.WaitVisible("#login", chromedp.ByID),
	)

	if err != nil {
		return nil, nil, err
	}

	err = chromedp.Run(ctx,
		chromedp.Sleep(1*time.Second),
		chromedp.WaitVisible("#username", chromedp.ByID),
		chromedp.Focus("#username", chromedp.ByID),
		chromedp.SendKeys("#username", "LSPD", chromedp.ByID),
		chromedp.Focus("#password", chromedp.ByID),
		chromedp.SendKeys("#password", Password, chromedp.ByID),
		chromedp.Click(`#login > fieldset > input.button2.specialbutton`, chromedp.ByQuery),
	)

	if err != nil {
		cancel()
		return nil, nil, err
	}

	return ctx, cancel, nil
}

func FetchUserGroups(name string, discord string) ([]string, error) {
	ctx, cancel, err := initScraper()
	if err != nil {
		return nil, err
	}
	defer cancel()

	parsedName := strings.Split(name, " ")
	forQuery := parsedName[0] + "+" + parsedName[1]

	var ret string
	err = chromedp.Run(ctx,
		chromedp.WaitVisible("#username_logged_in"),
		chromedp.Navigate(fmt.Sprintf(SearchPageURL, forQuery)),
	)

	if err != nil {
		return nil, err
	}

	if strings.Compare(ret, "No members found for this search criterion.") == 0 {
		return nil, errors.New("account not found on the forums")
	}

	class1 := "username-coloured"
	class2 := "username"
	var ok1, ok2 bool
	err = chromedp.Run(ctx,
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

	err = chromedp.Run(ctx, chromedp.AttributeValue(hrefSelector, "href", &href, nil))
	if err != nil {
		return nil, errors.New("couldn't get account's URL")
	}

	var forumDiscord string
	err = chromedp.Run(ctx,
		chromedp.Navigate("https://lspd.ro-rp.ro/"+href[2:]),
		chromedp.Text(DiscordNameSelector, &forumDiscord, chromedp.NodeVisible),
	)

	if err != nil {
		return nil, err
	}

	if strings.Compare(forumDiscord, discord) != 0 {
		return nil, errors.New("failed to check discord name on the forums")
	}

	var optionsGroup []*cdp.Node
	err = chromedp.Run(ctx, chromedp.Nodes("select option", &optionsGroup))
	if err != nil {
		return nil, err
	}

	var groupIds []string
	for _, groupNode := range optionsGroup {
		for _, group := range groupNode.Children {
			groupIds = append(groupIds, group.NodeValue)
		}
	}

	return groupIds, nil
}
