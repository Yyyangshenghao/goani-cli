package source

import "github.com/Yyyangshenghao/goani-cli/internal/settings"

// Subscription 订阅配置
type Subscription struct {
	URL       string `json:"url"`
	Name      string `json:"name"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

func defaultSubscriptions() []Subscription {
	return subscriptionsFromSettings(settings.DefaultSubscriptions())
}

func subscriptionsFromSettings(subs []settings.Subscription) []Subscription {
	if len(subs) == 0 {
		return nil
	}

	result := make([]Subscription, len(subs))
	for i, sub := range subs {
		result[i] = Subscription{
			URL:  sub.URL,
			Name: sub.Name,
		}
	}
	return result
}

func subscriptionsToSettings(subs []Subscription) []settings.Subscription {
	if len(subs) == 0 {
		return nil
	}

	result := make([]settings.Subscription, len(subs))
	for i, sub := range subs {
		result[i] = settings.Subscription{
			URL:  sub.URL,
			Name: sub.Name,
		}
	}
	return result
}

func cloneSubscriptions(subs []Subscription) []Subscription {
	if len(subs) == 0 {
		return nil
	}

	cloned := make([]Subscription, len(subs))
	copy(cloned, subs)
	return cloned
}

func sameSubscriptions(a, b []Subscription) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].URL != b[i].URL || a[i].Name != b[i].Name {
			return false
		}
	}
	return true
}
