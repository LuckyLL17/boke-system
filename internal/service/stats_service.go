package service

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"podcast-platform/config"
	"podcast-platform/internal/domain"
	"podcast-platform/internal/repository"
	appErr "podcast-platform/pkg/errors"
	"podcast-platform/pkg/ipgeo"
	"podcast-platform/pkg/utils"

	"gorm.io/gorm"
)

type StatsService struct {
	playbackRepo   *repository.PlaybackRepository
	episodeRepo    *repository.EpisodeRepository
	channelRepo    *repository.ChannelRepository
	subscriberRepo *repository.SubscriberRepository
	geoResolver    *ipgeo.Resolver
}

func NewStatsService(playbackRepo *repository.PlaybackRepository, episodeRepo *repository.EpisodeRepository,
	channelRepo *repository.ChannelRepository, subscriberRepo *repository.SubscriberRepository) *StatsService {
	return &StatsService{
		playbackRepo:   playbackRepo,
		episodeRepo:    episodeRepo,
		channelRepo:    channelRepo,
		subscriberRepo: subscriberRepo,
		geoResolver:    ipgeo.NewResolver(false),
	}
}

type DashboardStats struct {
	TotalPlays        int64                    `json:"total_plays"`
	TotalListeners    int64                    `json:"total_listeners"`
	TotalSubscribers  int64                    `json:"total_subscribers"`
	TotalEpisodes     int64                    `json:"total_episodes"`
	PlaysByDate       []DatePoint              `json:"plays_by_date"`
	ListenersByDate   []DatePoint              `json:"listeners_by_date"`
	SubscribersByDate []DatePoint              `json:"subscribers_by_date"`
	TopEpisodes       []domain.Episode         `json:"top_episodes"`
	ByCountry         []map[string]interface{} `json:"by_country"`
	ByDevice          []map[string]interface{} `json:"by_device"`
	ByHour            map[int]int64            `json:"by_hour"`
	ByDayOfWeek       map[int]int64            `json:"by_day_of_week"`
}

type DatePoint struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type RangeRequest struct {
	ChannelID uint64
	Range     string
	StartDate string
	EndDate   string
}

func (s *StatsService) parseRange(req RangeRequest) (time.Time, time.Time) {
	from, to := time.Time{}, time.Time{}
	now := utils.Now()
	switch req.Range {
	case "7d":
		from = utils.DaysAgo(7)
		to = utils.EndOfDay(now)
	case "30d":
		from = utils.DaysAgo(30)
		to = utils.EndOfDay(now)
	case "90d":
		from = utils.DaysAgo(90)
		to = utils.EndOfDay(now)
	case "month":
		from = utils.StartOfMonth(now)
		to = utils.EndOfDay(now)
	case "custom":
		if req.StartDate != "" {
			from, _ = utils.ParseDate(req.StartDate)
		}
		if req.EndDate != "" {
			to, _ = utils.ParseDate(req.EndDate)
			to = utils.EndOfDay(to)
		}
	default:
		from = utils.DaysAgo(30)
		to = utils.EndOfDay(now)
	}
	return from, to
}

func (s *StatsService) GetChannelDashboard(req RangeRequest) (*DashboardStats, error) {
	from, to := s.parseRange(req)
	stats := &DashboardStats{}

	plays, err := s.playbackRepo.CountByChannel(req.ChannelID, from, to)
	if err == nil {
		stats.TotalPlays = plays
	}
	listeners, err := s.playbackRepo.CountUniqueListeners(req.ChannelID, from, to)
	if err == nil {
		stats.TotalListeners = listeners
	}
	subs, err := s.subscriberRepo.CountByChannel(req.ChannelID)
	if err == nil {
		stats.TotalSubscribers = subs
	}
	_, eps, err := s.episodeRepo.ListByChannel(req.ChannelID, 1, 1, nil, "")
	if err == nil {
		stats.TotalEpisodes = eps
	}

	playsByDate, _ := s.playbackRepo.GroupByDate(req.ChannelID, from, to)
	stats.PlaysByDate = buildDatePoints(playsByDate, from, to, "count")
	stats.ListenersByDate = buildDatePoints(playsByDate, from, to, "listeners")

	subsByDate, _ := s.subscriberRepo.GroupByDate(req.ChannelID, from, to)
	stats.SubscribersByDate = buildDatePoints(subsByDate, from, to, "count")

	topEps, _ := s.episodeRepo.GetTopByChannel(req.ChannelID, 10)
	stats.TopEpisodes = topEps

	countryData, _ := s.playbackRepo.GroupByCountry(req.ChannelID, from, to)
	stats.ByCountry = countryData

	deviceData, _ := s.playbackRepo.GroupByDevice(req.ChannelID, from, to)
	stats.ByDevice = deviceData

	hourData, _ := s.playbackRepo.GroupByHour(req.ChannelID, from, to)
	stats.ByHour = buildIntMap(hourData, "hour", "count")

	dowData, _ := s.playbackRepo.GroupByDayOfWeek(req.ChannelID, from, to)
	stats.ByDayOfWeek = buildIntMap(dowData, "dow", "count")

	return stats, nil
}

func buildDatePoints(data []map[string]interface{}, from, to time.Time, key string) []DatePoint {
	points := make([]DatePoint, 0)
	if from.IsZero() || to.IsZero() {
		return points
	}
	dataMap := make(map[string]int64)
	for _, d := range data {
		date := ""
		if v, ok := d["date"].(time.Time); ok {
			date = v.Format("2006-01-02")
		} else if s, ok := d["date"].(string); ok {
			date = s
		}
		count := toInt64(d[key])
		if date != "" {
			dataMap[date] = count
		}
	}
	for d := utils.StartOfDay(from); !d.After(to); d = d.AddDate(0, 0, 1) {
		ds := d.Format("2006-01-02")
		points = append(points, DatePoint{Date: ds, Count: dataMap[ds]})
	}
	return points
}

func buildIntMap(data []map[string]interface{}, keyField, valField string) map[int]int64 {
	result := make(map[int]int64)
	for _, d := range data {
		k := toInt(d[keyField])
		v := toInt64(d[valField])
		result[k] = v
	}
	return result
}

func toInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	case string:
		n, _ := strconv.Atoi(val)
		return n
	}
	return 0
}

func toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case int:
		return int64(val)
	case int64:
		return val
	case float64:
		return int64(val)
	case string:
		n, _ := strconv.ParseInt(val, 10, 64)
		return n
	}
	return 0
}

type PlaybackStartRequest struct {
	EpisodeID uint64
	UserID    uint64
	IPAddress string
	UserAgent string
	Source    string
}

func (s *StatsService) RecordPlaybackStart(req PlaybackStartRequest) (uint64, error) {
	ep, err := s.episodeRepo.GetByID(req.EpisodeID)
	if err != nil {
		return 0, appErr.ErrEpisodeNotFound
	}
	device, os, browser := ipgeo.ParseUserAgent(req.UserAgent)
	loc := s.geoResolver.Resolve(req.IPAddress)
	pb := &domain.Playback{
		EpisodeID:  req.EpisodeID,
		ChannelID:  ep.ChannelID,
		UserID:     req.UserID,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
		DeviceType: device,
		OS:         os,
		Browser:    browser,
		Country:    loc.Country,
		Region:     loc.Region,
		City:       loc.City,
		StartAt:    utils.Now(),
		Source:     req.Source,
	}
	err = s.playbackRepo.DB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(pb).Error; err != nil {
			return err
		}
		if err := s.episodeRepo.WithTx(tx).IncrementPlayCount(req.EpisodeID, 1); err != nil {
			return err
		}
		if err := s.channelRepo.WithTx(tx).IncrementTotalPlays(ep.ChannelID, 1); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return 0, appErr.Wrap(err, 500, "record playback")
	}
	return pb.ID, nil
}

type PlaybackEndRequest struct {
	PlaybackID uint64
	ListenTime int
	Progress   float64
	Completed  bool
}

func (s *StatsService) RecordPlaybackEnd(req PlaybackEndRequest) error {
	if req.PlaybackID == 0 {
		return nil
	}
	type raw struct {
		ID    uint64
		Start time.Time
	}
	var r raw
	_ = r
	return nil
}

func (s *StatsService) GetEpisodeCompletion(episodeID uint64) (float64, int64, error) {
	return s.playbackRepo.CompletionRate(episodeID)
}

func (s *StatsService) ExportCSV(channelID uint64, from, to time.Time, w http.ResponseWriter) error {
	plays, _ := s.playbackRepo.GroupByDate(channelID, from, to)
	subs, _ := s.subscriberRepo.GroupByDate(channelID, from, to)
	subsMap := make(map[string]int64)
	for _, d := range subs {
		date := ""
		if v, ok := d["date"].(time.Time); ok {
			date = v.Format("2006-01-02")
		} else if s, ok := d["date"].(string); ok {
			date = s
		}
		subsMap[date] = toInt64(d["count"])
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=stats.csv")
	writer := csvWriter{w: w}
	writer.writeRow([]string{"Date", "Plays", "Listeners", "New Subscribers"})
	for _, d := range plays {
		date := ""
		if v, ok := d["date"].(time.Time); ok {
			date = v.Format("2006-01-02")
		} else if s, ok := d["date"].(string); ok {
			date = s
		}
		playsCount := toInt64(d["count"])
		listeners := toInt64(d["listeners"])
		newSubs := subsMap[date]
		writer.writeRow([]string{date, strconv.FormatInt(playsCount, 10),
			strconv.FormatInt(listeners, 10), strconv.FormatInt(newSubs, 10)})
	}
	return nil
}

type csvWriter struct {
	w http.ResponseWriter
}

func (c *csvWriter) writeRow(cols []string) {
	c.w.Write([]byte(strings.Join(cols, ",") + "\n"))
}

func (s *StatsService) AggregateDailyStats() (int, error) {
	ids, err := s.channelRepo.GetAllIDs(1000)
	if err != nil {
		return 0, err
	}
	today := utils.Now()
	count := 0
	for _, id := range ids {
		from := utils.StartOfDay(today.AddDate(0, 0, -1))
		to := utils.EndOfDay(today.AddDate(0, 0, -1))
		plays, _ := s.playbackRepo.CountByChannel(id, from, to)
		listeners, _ := s.playbackRepo.CountUniqueListeners(id, from, to)
		subs, _ := s.subscriberRepo.CountByChannelDate(id, from, to)
		data := map[string]interface{}{
			"plays":     plays,
			"listeners": listeners,
			"subs":      subs,
		}
		b, _ := json.Marshal(data)
		cache := &domain.StatsCache{
			ChannelID:   id,
			Date:        from,
			PlayCount:   plays,
			Listeners:   int(listeners),
			Subscribers: int(subs),
			Data:        string(b),
		}
		if err := s.playbackRepo.SaveStatsCache(cache); err == nil {
			count++
		}
	}
	return count, nil
}

func (s *StatsService) GetTTL() int {
	return config.AppConfig.Cache.StatsTTL
}

type DailySummaryRow = repository.DailySummaryRow

func (s *StatsService) UpdateDailyCache(channelID uint64, date time.Time, row DailySummaryRow) error {
	data := map[string]interface{}{
		"playbacks":        row.TotalPlaybacks,
		"unique_users":     row.UniqueUsers,
		"total_listen_sec": row.TotalListenSec,
	}
	b, _ := json.Marshal(data)
	cache := &domain.StatsCache{
		ChannelID:   channelID,
		Date:        date,
		PlayCount:   row.TotalPlaybacks,
		Listeners:   int(row.UniqueUsers),
		Subscribers: 0,
		Data:        string(b),
	}
	existing, err := s.playbackRepo.GetStatsCache(channelID, date)
	if err == nil && existing != nil {
		cache.ID = existing.ID
	}
	return s.playbackRepo.SaveStatsCache(cache)
}
