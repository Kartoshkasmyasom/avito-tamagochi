package tasks

func taskText(a string) (string, string) {
	switch a {
	case ActivityListingViewed:
		return "Посмотреть объявление", "Посмотрите объявление на Авито"
	case ActivityFavoriteAdded:
		return "Добавить в избранное", "Добавьте объявление в избранное"
	default:
		return "Опубликовать объявление", "Опубликуйте объявление на Авито"
	}
}

func Title(activity string) string {
	title, _ := taskText(activity)
	return title
}

type definition struct {
	activity   string
	target, xp int
}

func definitions() []definition {
	return []definition{{ActivityListingViewed, 1, 40}, {ActivityFavoriteAdded, 1, 70}, {ActivityListingPublished, 1, 100}}
}
