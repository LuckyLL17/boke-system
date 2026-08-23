package repository

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"podcast-platform/internal/domain"
)

// ChannelHandler.Subscribe -> SubscriberService.SourceDistribution -> SubscriberRepository.GroupBySource
func TestSourceDistributionOmitsUnattributedActiveRows(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:task019?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.Subscriber{}); err != nil {
		t.Fatal(err)
	}
	rows := []domain.Subscriber{
		{ChannelID: 3, Email: "blank@example.com", Status: 1, Source: ""},
		{ChannelID: 3, Email: "campaign@example.com", Status: 1, Source: "campaign"},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatal(err)
	}
	groups, err := NewSubscriberRepository(db).GroupBySource(3)
	if err != nil {
		t.Fatal(err)
	}
	for _, group := range groups {
		if group["source"] == "" {
			t.Fatalf("unattributed active subscriber leaked into source distribution: %#v", groups)
		}
	}
}
