package jparser_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/egelis/jparser"
)

type args struct {
	data json.RawMessage
	meta []jparser.MetaData
}

func TestParseParamsSuccess(t *testing.T) {
	testTable := []struct {
		name        string
		args        args
		expectedRes []jparser.RawMessageSet
	}{
		{
			name: "JSON with one element in array",
			args: args{
				data: oneElementInArrayJSON,
				meta: []jparser.MetaData{
					{"[].UL.branches.[].kpp", "kpp"},
					{"[].inn", "inn"},
					{"[].UL.branches.[].date", "date"},
					{"[].UL.legalAddress.parsedAddressRF.non-existing", "non-existing"},
				},
			},
			expectedRes: []jparser.RawMessageSet{
				{
					"date": json.RawMessage(`"2008-10-03"`),
					"inn":  json.RawMessage(`"6663003127"`),
					"kpp":  json.RawMessage(`"771543001"`),
				},
				{
					"date": json.RawMessage(`"2011-09-02"`),
					"inn":  json.RawMessage(`"6663003127"`),
					"kpp":  json.RawMessage(`"771543002"`),
				},
				{
					"date": json.RawMessage(`"2017-11-22"`),
					"inn":  json.RawMessage(`"6663003127"`),
					"kpp":  json.RawMessage(`"780243001"`),
				},
				{
					"date": json.RawMessage(`"2018-05-24"`),
					"inn":  json.RawMessage(`"6663003127"`),
					"kpp":  json.RawMessage(`"590443001"`),
				},
				{
					"date": json.RawMessage(`"2021-09-09"`),
					"inn":  json.RawMessage(`"6663003127"`),
					"kpp":  json.RawMessage(`"745343002"`),
				},
			},
		},
		{
			name: "JSON with multiple elements in array",
			args: args{
				data: multipleElementsInArrayJSON,
				meta: []jparser.MetaData{
					{"[].UL.branches.[].date", "date1"},
					{"[].IP.status.date", "date2"},
				},
			},
			expectedRes: []jparser.RawMessageSet{
				{},
				{
					"date2": json.RawMessage(`"2017-05-05"`),
				},
				{
					"date2": json.RawMessage(`"2013-03-13"`),
				},
			},
		},
		{
			name: "JSON with object",
			args: args{
				data: multipleElementsInArrayJSON,
				meta: []jparser.MetaData{},
			},
			expectedRes: []jparser.RawMessageSet{{}},
		},
		{
			name: "Empty JSON",
			args: args{
				data: oneObjectInJSON,
				meta: []jparser.MetaData{
					{"inn", "inn"},
					{"IP.status.statusString", "status"},
				},
			},
			expectedRes: []jparser.RawMessageSet{
				{
					"inn":    json.RawMessage(`"772473497153"`),
					"status": json.RawMessage(`"Действующее"`),
				},
			},
		},
		{
			name: "Empty array in JSON",
			args: args{
				data: json.RawMessage(`[]`),
				meta: []jparser.MetaData{
					{"[].UL.heads", "heads"},
				},
			},
			expectedRes: []jparser.RawMessageSet{{}},
		},
		{
			name: "Test # and @ symbols",
			args: args{
				data: oneElementInArrayJSON,
				meta: []jparser.MetaData{
					{"[].UL.branches.[].@", "branches_index"},
					{"[].UL.branches.[].#", "branches_count"},
				},
			},
			expectedRes: []jparser.RawMessageSet{
				{
					"branches_index": json.RawMessage(`0`),
					"branches_count": json.RawMessage(`5`),
				},
				{
					"branches_index": json.RawMessage(`1`),
					"branches_count": json.RawMessage(`5`),
				},
				{
					"branches_index": json.RawMessage(`2`),
					"branches_count": json.RawMessage(`5`),
				},
				{
					"branches_index": json.RawMessage(`3`),
					"branches_count": json.RawMessage(`5`),
				},
				{
					"branches_index": json.RawMessage(`4`),
					"branches_count": json.RawMessage(`5`),
				},
			},
		},
		{
			name: "Get array from JSON",
			args: args{
				data: oneElementInArrayJSON,
				meta: []jparser.MetaData{
					{"[].UL.history.kpps.[]", "kpps"},
				},
			},
			expectedRes: []jparser.RawMessageSet{
				{
					"kpps": json.RawMessage(`[
                    {
                        "kpp": "668601001",
                        "date": "2016-11-19"
                    },
                    {
                        "kpp": "667301001",
                        "date": "2005-07-29"
                    }
                ]`),
				},
			},
		},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			result, err := jparser.ParseParams(test.args.data, test.args.meta)

			if err != nil {
				t.Errorf("ParseParams() got error = \"%v\", expected nil", err)
				return
			}

			if !reflect.DeepEqual(result, test.expectedRes) {
				got, _ := json.MarshalIndent(result, "", "  ")
				expected, _ := json.MarshalIndent(test.expectedRes, "", "  ")
				t.Errorf("ParseParams() got result = %s\nexpectedRes = %s", got, expected)
			}
		})
	}
}

func TestParseParamsErrors(t *testing.T) {
	testTable := []struct {
		name string
		args args
	}{
		{
			name: "JSON is broken",
			args: args{
				data: brokenJSON,
				meta: []jparser.MetaData{
					{"[].inn", "inn"},
				},
			},
		},
		{
			name: "For existing JSON path we have wrong meta (JSON has '[]' in path, meta has 'object')",
			args: args{
				data: oneElementInArrayJSON,
				meta: []jparser.MetaData{
					{"[].UL.branches.[].kpp", "kpp_param"},
					{"[].UL.branches.wrong_path", "wrong_path_param"},
				},
			},
		},
		{
			name: "For existing JSON path we have wrong meta (JSON has 'object' in path, meta has '[]')",
			args: args{
				data: oneElementInArrayJSON,
				meta: []jparser.MetaData{
					{"[].UL.branches.[].kpp", "kpp"},
					{"[].UL.[].wrong_path", "wrong_path"},
				},
			},
		},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			result, err := jparser.ParseParams(test.args.data, test.args.meta)

			if err == nil {
				t.Errorf("ParseParams() got error = nil, expected error")
				return
			}

			if result != nil {
				got, _ := json.MarshalIndent(result, "", "  ")
				t.Errorf("ParseParams() got result = %s, expectedRes = nil", got)
			}
		})
	}
}

var (
	oneObjectInJSON = json.RawMessage(`
{
    "inn": "772473497153",
    "ogrn": "318774600372150",
    "focusHref": "https://focus.kontur.ru/entity?query=318774600372150",
    "IP": {
        "fio": "Щербина Илья Владимирович",
        "okpo": "0133585313",
        "okato": "45296590000",
        "okfs": "16",
        "okogu": "4210015",
        "okopf": "50102",
        "opf": "Индивидуальные предприниматели",
        "oktmo": "45923000000",
        "registrationDate": "2018-07-11",
        "status": {
            "statusString": "Действующее"
        }
    },
    "briefReport": {
        "summary": {
            "greenStatements": true
        }
    },
    "contactPhones": {}
}
`)

	oneElementInArrayJSON = json.RawMessage(`
[
    {
        "inn": "6663003127",
        "ogrn": "1026605606620",
        "focusHref": "https://focus.kontur.ru/entity?query=1026605606620",
        "UL": {
            "kpp": "667101001",
            "okpo": "00242766",
            "okato": "65401377000",
            "okfs": "16",
            "oktmo": "65701000001",
            "okogu": "4210014",
            "okopf": "12267",
            "opf": "Непубличные акционерные общества",
            "legalName": {
                "short": "АО \"ПФ \"СКБ Контур\"",
                "full": "Акционерное общество \"Производственная Фирма \"СКБ Контур\"",
                "readable": "АО \"ПФ \"СКБ Контур\"",
                "date": "2017-06-21"
            },
            "legalAddress": {
                "parsedAddressRF": {
                    "zipCode": "620144",
                    "kladrCode": "660000010000717",
                    "regionCode": "66",
                    "regionName": {
                        "topoShortName": "обл",
                        "topoFullName": "область",
                        "topoValue": "Свердловская"
                    },
                    "city": {
                        "topoShortName": "г",
                        "topoFullName": "город",
                        "topoValue": "Екатеринбург"
                    },
                    "street": {
                        "topoShortName": "ул",
                        "topoFullName": "улица",
                        "topoValue": "Народной воли"
                    },
                    "bulk": {
                        "topoShortName": "стр",
                        "topoFullName": "строение",
                        "topoValue": "19А"
                    },
                    "bulkRaw": "СТР. 19А"
                },
                "date": "2020-09-02",
                "parsedAddressRFFias": {
                    "fiasId": 46909691,
                    "fiasGUID": "c47cb575-18cf-47a7-9cdd-6c750d89657a",
                    "zipCode": "620144",
                    "regionCode": "66",
                    "region": {
                        "topoShortName": "обл.",
                        "topoFullName": "область",
                        "topoValue": "Свердловская"
                    },
                    "municipalDistrict": {
                        "topoShortName": "г.о.",
                        "topoFullName": "городской округ",
                        "topoValue": "город Екатеринбург"
                    },
                    "city": {
                        "topoShortName": "г.",
                        "topoFullName": "город",
                        "topoValue": "Екатеринбург"
                    },
                    "street": {
                        "topoShortName": "ул.",
                        "topoFullName": "улица",
                        "topoValue": "Народной воли"
                    },
                    "buildings": [
                        {
                            "topoShortName": "стр.",
                            "topoFullName": "строение",
                            "topoValue": "19а"
                        }
                    ],
                    "isConverted": true
                },
                "firstDate": "2020-09-02"
            },
            "branches": [
                {
                    "kpp": "771543001",
                    "parsedAddressRF": {
                        "zipCode": "127254",
                        "kladrCode": "770000000004162",
                        "regionCode": "77",
                        "regionName": {
                            "topoShortName": "г",
                            "topoFullName": "город",
                            "topoValue": "Москва"
                        },
                        "street": {
                            "topoShortName": "пер",
                            "topoFullName": "переулок",
                            "topoValue": "Добролюбова"
                        },
                        "house": {
                            "topoShortName": "дом",
                            "topoFullName": "дом",
                            "topoValue": "3"
                        },
                        "bulk": {
                            "topoShortName": "корп",
                            "topoFullName": "корпус",
                            "topoValue": "1"
                        },
                        "flat": {
                            "topoShortName": "кв",
                            "topoFullName": "квартира",
                            "topoValue": "412"
                        },
                        "houseRaw": "Д.3",
                        "bulkRaw": "К.1",
                        "flatRaw": "КВ.412"
                    },
                    "date": "2008-10-03"
                },
                {
                    "kpp": "771543002",
                    "parsedAddressRF": {
                        "zipCode": "127018",
                        "kladrCode": "770000000000606",
                        "regionCode": "77",
                        "regionName": {
                            "topoShortName": "г",
                            "topoFullName": "город",
                            "topoValue": "Москва"
                        },
                        "street": {
                            "topoShortName": "ул",
                            "topoFullName": "улица",
                            "topoValue": "Сущевский Вал"
                        },
                        "house": {
                            "topoShortName": "дом",
                            "topoFullName": "дом",
                            "topoValue": "18"
                        },
                        "houseRaw": "Д.18"
                    },
                    "date": "2011-09-02"
                },
                {
                    "kpp": "780243001",
                    "parsedAddressRF": {
                        "zipCode": "194044",
                        "kladrCode": "780000000000292",
                        "regionCode": "78",
                        "regionName": {
                            "topoShortName": "г",
                            "topoFullName": "город",
                            "topoValue": "Санкт-Петербург"
                        },
                        "street": {
                            "topoShortName": "ул",
                            "topoFullName": "улица",
                            "topoValue": "Гельсингфорсская"
                        },
                        "house": {
                            "topoShortName": "дом",
                            "topoFullName": "дом",
                            "topoValue": "2"
                        },
                        "bulk": {
                            "topoShortName": "лит",
                            "topoFullName": "литера",
                            "topoValue": "А"
                        },
                        "flat": {
                            "topoValue": "помещ. 11н, 12н"
                        },
                        "houseRaw": "Д. 2",
                        "bulkRaw": "ЛИТЕР А",
                        "flatRaw": "ПОМЕЩ. 11Н, 12Н"
                    },
                    "date": "2017-11-22"
                },
                {
                    "name": "ПЕРМСКИЙ ФИЛИАЛ АО \"ПФ \"СКБ КОНТУР\"",
                    "kpp": "590443001",
                    "parsedAddressRF": {
                        "zipCode": "614010",
                        "kladrCode": "590000010000603",
                        "regionCode": "59",
                        "regionName": {
                            "topoShortName": "край",
                            "topoFullName": "край",
                            "topoValue": "Пермский"
                        },
                        "city": {
                            "topoShortName": "г",
                            "topoFullName": "город",
                            "topoValue": "Пермь"
                        },
                        "street": {
                            "topoShortName": "ул",
                            "topoFullName": "улица",
                            "topoValue": "Куйбышева"
                        },
                        "house": {
                            "topoShortName": "дом",
                            "topoFullName": "дом",
                            "topoValue": "95"
                        },
                        "bulk": {
                            "topoShortName": "корп",
                            "topoFullName": "корпус",
                            "topoValue": "Б"
                        },
                        "flat": {
                            "topoShortName": "пом",
                            "topoFullName": "помещение",
                            "topoValue": "1"
                        },
                        "houseRaw": "Д. 95",
                        "bulkRaw": "К. Б",
                        "flatRaw": "ПОМЕЩ. 1"
                    },
                    "date": "2018-05-24"
                },
                {
                    "kpp": "745343002",
                    "parsedAddressRF": {
                        "zipCode": "454080",
                        "kladrCode": "740000010000168",
                        "regionCode": "74",
                        "regionName": {
                            "topoShortName": "обл",
                            "topoFullName": "область",
                            "topoValue": "Челябинская"
                        },
                        "city": {
                            "topoShortName": "г",
                            "topoFullName": "город",
                            "topoValue": "Челябинск"
                        },
                        "street": {
                            "topoShortName": "ул",
                            "topoFullName": "улица",
                            "topoValue": "Витебская"
                        },
                        "house": {
                            "topoShortName": "дом",
                            "topoFullName": "дом",
                            "topoValue": "4"
                        },
                        "houseRaw": "Д. 4",
                        "isConverted": true
                    },
                    "date": "2021-09-09"
                }
            ],
            "status": {
                "statusString": "Действующее"
            },
            "registrationDate": "1992-03-26",
            "history": {
                "kpps": [
                    {
                        "kpp": "668601001",
                        "date": "2016-11-19"
                    },
                    {
                        "kpp": "667301001",
                        "date": "2005-07-29"
                    }
                ]
            }
        },
        "briefReport": {
            "summary": {
                "greenStatements": true
            }
        },
        "contactPhones": {
            "count": 77
        }
    }
]
`)

	multipleElementsInArrayJSON = json.RawMessage(`
[
    {
        "inn": "772473497153",
        "ogrn": "318774600372150",
        "focusHref": "https://focus.kontur.ru/entity?query=318774600372150",
        "IP": {
            "fio": "Щербина Илья Владимирович",
            "okpo": "0133585313",
            "okato": "45296590000",
            "okfs": "16",
            "okogu": "4210015",
            "okopf": "50102",
            "opf": "Индивидуальные предприниматели",
            "oktmo": "45923000000",
            "registrationDate": "2018-07-11",
            "status": {
                "statusString": "Действующее"
            }
        },
        "briefReport": {
            "summary": {
                "greenStatements": true
            }
        },
        "contactPhones": {}
    },
    {
        "inn": "772473497153",
        "ogrn": "314774614000310",
        "focusHref": "https://focus.kontur.ru/entity?query=314774614000310",
        "IP": {
            "fio": "Щербина Илья Владимирович",
            "okpo": "0116259884",
            "okato": "45296590000",
            "okfs": "16",
            "okogu": "4210015",
            "okopf": "50102",
            "opf": "Индивидуальные предприниматели",
            "registrationDate": "2014-05-20",
            "dissolutionDate": "2017-05-05",
            "status": {
                "statusString": "Индивидуальный предприниматель прекратил деятельность в связи с принятием им соответствующего решения",
                "dissolved": true,
                "date": "2017-05-05"
            }
        },
        "briefReport": {
            "summary": {
                "redStatements": true
            }
        },
        "contactPhones": {}
    },
    {
        "inn": "772473497153",
        "ogrn": "307770000117071",
        "focusHref": "https://focus.kontur.ru/entity?query=307770000117071",
        "IP": {
            "fio": "Щербина Илья Владимирович",
            "okpo": "0116259884",
            "okato": "45296590000",
            "okogu": "49015",
            "okopf": "50102",
            "opf": "Индивидуальные предприниматели",
            "registrationDate": "2007-03-06",
            "dissolutionDate": "2013-03-13",
            "status": {
                "statusString": "Индивидуальный предприниматель прекратил деятельность в связи с принятием им соответствующего решения",
                "dissolved": true,
                "date": "2013-03-13"
            }
        },
        "briefReport": {
            "summary": {
                "redStatements": true
            }
        },
        "contactPhones": {}
    }
]
`)

	brokenJSON = json.RawMessage(`
[
    {
        "inn": "7452160483",
        "ogrn": "1227400033629",
        "focusHref": "https://focus.kontur.ru/entity?query=1227400033629",
        : {
            "okpo": "57422264",
            "pfrRegNumber": "084004086979",
            "fssRegNumber": "740204225274021",
            "activities": {
                "principalActivity": {
                    "code": "43.21",
                    "text": "Производство электромонтажных работ",
                    "date": "2022-09-07"
                },
                "complementaryActivities": [
                    {
                        "code": "25.61",
                        "text": "Обработка металлов и нанесение покрытий на металлы",
                        "date": "2022-09-07"
                    },
                    {
                        "code": "25.62",
                        "text": "Обработка металлических изделий механическая",
                        "date": "2022-09-07"
                    },
                    {
                        "code": "38.32.2",
                        "text": "Обработка отходов и лома драгоценных металлов",
                        "date": "2022-09-07"
                    },
                ],
                "okvedVersion": "2"
            },
        }
    }
]
`)
)
