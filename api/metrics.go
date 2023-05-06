package api

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func (t TallyServer) collectStatsAtInterval(interval time.Duration) {
	t.collectStats()
	time.AfterFunc(interval, func() { t.collectStatsAtInterval(interval) })
}
func (t TallyServer) collectStats() {
	stats, err := t.storage.Stats(context.TODO())
	if err != nil {
		t.l.Error().Err(err).Msg("failed to collect stats")
		return
	}
	metricTotalUsers.Set(float64(stats.Users))
	metricTotalSessions.Set(float64(stats.Session))
	metricTotalGames.Set(float64(stats.Games))
	metricTotalGamesWon.Set(float64(stats.GamesWon))
	metricTotalGamesLost.Set(float64(stats.GamesLost))
	metricTotalGamesAbandoned.Set(float64(stats.GamesAbandoned))
	metricTotalGamesCurrent.Set(float64(stats.GamesCurrent))
	metricLongestGame.Set(float64(stats.LongestGame))
	metricHighestScore.Set(float64(stats.HighestScore))
	metricHistoryStdDev.Set(float64(stats.HistoryStdDev))
	metricHistoryMin.Set(float64(stats.HistoryMin))
	metricHistoryAvg.Set(float64(stats.HistoryAvg))
	metricHistoryMax.Set(float64(stats.HistoryMax))
	metricHistoryTotal.Set(float64(stats.HistoryTotal))
}

var (
	metricTotalUsers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gotally_users_total",
		Help: "The total number of registered users",
	})
	metricTotalSessions = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gotally_sessions_total",
		Help: "The total number of registered sessions",
	})
	metricTotalGames = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gotally_games_total",
		Help: "The total number of recorded games",
	})
	metricTotalGamesWon = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gotally_games_won_total",
		Help: "The total number of recorded games won",
	})
	metricTotalGamesLost = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gotally_games_lost_total",
		Help: "The total number of recorded games lost",
	})
	metricTotalGamesAbandoned = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gotally_games_abandoned_total",
		Help: "The total number of recorded games adandoned",
	})
	metricTotalGamesCurrent = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gotally_games_current_total",
		Help: "The total number of recorded games current",
	})
	metricLongestGame = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gotally_game_longest",
		Help: "The number of moves in the longest game",
	})
	metricHighestScore = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gotally_game_highestscore",
		Help: "Highest score achived",
	})
	metricHistoryStdDev = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gotally_game_history_size_std_dev",
		Help: "Size of data in game-history, represented as standard deviation",
	})
	metricHistoryAvg = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gotally_game_history_size_avg",
		Help: "Size of data in game-history, represented as average",
	})
	metricHistoryMax = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gotally_game_history_size_max",
		Help: "Size of data in game-history, represented as max",
	})
	metricHistoryMin = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gotally_game_history_size_min",
		Help: "Size of data in game-history, represented as min",
	})
	metricHistoryTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gotally_game_history_size_total",
		Help: "Size of data in game-history, represented as a total",
	})

	metricHttpCalls = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gotally_http_calls",
		Help: "Size of data in game-history, represented as a total",
	}, []string{"method", "path", "code"})
)
