package models

type QbitorrentServerState struct {
	Alltime_dl             int
	Alltime_ul             int
	Average_time_queue     int
	Connection_status      string
	Dht_nodes              int
	Dl_info_data           int
	Dl_info_speed          int
	Dl_rate_limit          int
	Free_space_on_disk     int
	Global_ratio           float64
	Queued_io_jobs         int
	Queueing               bool
	Read_cache_hits        float64
	Read_cache_overload    int
	Refresh_interval       int
	Total_buffers_size     int
	Total_peer_connections int
	Total_queued_size      int
	Total_wasted_session   int
	Up_info_data           int
	Up_info_speed          int
	Up_rate_limit          int
	Use_alt_speed_limits   bool
	Use_subcategories      bool
	Write_cache_overload   int
}

type QbitorrentTorrentData struct {
	Added_on                    int
	Amount_left                 int
	Auto_tmm                    bool
	Availability                int
	Category                    string
	Completed                   int
	Completion_on               int
	Content_path                string
	Dl_limit                    int
	Dlspeed                     int
	Download_path               string
	Downloaded                  int
	Downloaded_session          int
	Eta                         int
	F_l_piece_prio              bool
	Force_start                 bool
	Inactive_seeding_time_limit int
	Infohash_v1                 string
	Infohash_v2                 string
	Last_activity               int
	Magnet_uri                  string
	Max_inactive_seeding_time   int
	Max_ratio                   int
	Max_seeding_time            int
	Name                        string
	Num_complete                int
	Num_incomplete              int
	Num_leechs                  int
	Num_seeds                   int
	Priority                    int
	Progress                    int
	Ratio                       float64
	Ratio_limit                 int
	Save_path                   string
	Seeding_time                int
	Seeding_time_limit          int
	Seen_complete               int
	Seq_dl                      bool
	Size                        int
	State                       string
	Super_seeding               bool
	Tags                        string
	Time_active                 int
	Total_size                  int
	Tracker                     string
	Trackers_count              int
	Up_limit                    int
	Uploaded                    int
	Uploaded_session            int
	Upspeed                     int
}

type Tracker []string

type QbitorrentGlobalData struct {
	Server_state QbitorrentServerState
	Torrents     map[string]QbitorrentTorrentData
	Trackers     map[string]Tracker
}

type QbitorrentRenderData struct {
	QbitorrentGlobalData QbitorrentGlobalData
	Metadata             ModuleMetada
}
