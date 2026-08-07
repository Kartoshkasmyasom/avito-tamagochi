package rewards

func rewardText(t string) (string, string) {
	if t == "listingPromotion" {
		return "Продвижение объявления", "Бонус на продвижение объявления"
	}
	return "Авито Доставка", "Бонус на Авито Доставку"
}

func Title(t string) string {
	title, _ := rewardText(t)
	return title
}
