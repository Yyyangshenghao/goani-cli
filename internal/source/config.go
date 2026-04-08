package source

// Config 订阅源配置
type Config struct {
    ExportedMediaSourceDataList struct {
        MediaSources []MediaSource `json:"mediaSources"`
    } `json:"exportedMediaSourceDataList"`
}

// MediaSource 媒体源
type MediaSource struct {
    FactoryID string   `json:"factoryId"`
    Version   int      `json:"version"`
    Arguments Arguments `json:"arguments"`
}

// Arguments 媒体源参数
type Arguments struct {
    Name        string      `json:"name"`
    Description string      `json:"description"`
    IconURL     string      `json:"iconUrl"`
    SearchConfig SearchConfig `json:"searchConfig"`
    Tier        int         `json:"tier"`
}

// SearchConfig 搜索配置
type SearchConfig struct {
    SearchURL               string `json:"searchUrl"`
    SearchUseOnlyFirstWord  bool   `json:"searchUseOnlyFirstWord"`
    SearchRemoveSpecial     bool   `json:"searchRemoveSpecial"`
    SubjectFormatID         string `json:"subjectFormatId"`
    ChannelFormatID         string `json:"channelFormatId"`
    DefaultResolution       string `json:"defaultResolution"`
    DefaultSubtitleLanguage string `json:"defaultSubtitleLanguage"`

    // 搜索结果选择器
    SelectorSubjectFormatA      SelectorSubjectFormatA      `json:"selectorSubjectFormatA"`
    SelectorSubjectFormatIndexed SelectorSubjectFormatIndexed `json:"selectorSubjectFormatIndexed"`

    // 剧集选择器
    SelectorChannelFormatFlattened  SelectorChannelFormatFlattened  `json:"selectorChannelFormatFlattened"`
    SelectorChannelFormatNoChannel  SelectorChannelFormatNoChannel  `json:"selectorChannelFormatNoChannel"`

    // 视频匹配
    MatchVideo MatchVideo `json:"matchVideo"`
}

// SelectorSubjectFormatA 格式A的选择器
type SelectorSubjectFormatA struct {
    SelectLists       string `json:"selectLists"`
    PreferShorterName bool   `json:"preferShorterName"`
}

// SelectorSubjectFormatIndexed 索引格式的选择器
type SelectorSubjectFormatIndexed struct {
    SelectNames       string `json:"selectNames"`
    SelectLinks       string `json:"selectLinks"`
    PreferShorterName bool   `json:"preferShorterName"`
}

// SelectorChannelFormatFlattened 扁平化频道格式
type SelectorChannelFormatFlattened struct {
    SelectChannelNames      string `json:"selectChannelNames"`
    MatchChannelName        string `json:"matchChannelName"`
    SelectEpisodeLists      string `json:"selectEpisodeLists"`
    SelectEpisodesFromList  string `json:"selectEpisodesFromList"`
    SelectEpisodeLinksFromList string `json:"selectEpisodeLinksFromList"`
    MatchEpisodeSortFromName string `json:"matchEpisodeSortFromName"`
}

// SelectorChannelFormatNoChannel 无频道格式
type SelectorChannelFormatNoChannel struct {
    SelectEpisodes          string `json:"selectEpisodes"`
    SelectEpisodeLinks      string `json:"selectEpisodeLinks"`
    MatchEpisodeSortFromName string `json:"matchEpisodeSortFromName"`
}

// MatchVideo 视频匹配配置
type MatchVideo struct {
    EnableNestedURL  bool              `json:"enableNestedUrl"`
    MatchNestedURL   string            `json:"matchNestedUrl"`
    MatchVideoURL    string            `json:"matchVideoUrl"`
    Cookies          string            `json:"cookies"`
    AddHeadersToVideo map[string]string `json:"addHeadersToVideo"`
}
