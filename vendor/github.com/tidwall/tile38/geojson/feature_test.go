package geojson

import "testing"

func TestFeature(t *testing.T) {
	testJSON(t, `{
    "type": "Feature",
    "geometry": {
        "type": "Polygon",
        "coordinates": [
            [
                [-80.72487831115721,35.26545403190955],
                [-80.72135925292969,35.26727607954368],
                [-80.71517944335938,35.26769654625573],
                [-80.7125186920166,35.27035945142482],
                [-80.70857048034668,35.268257165144064],
                [-80.70479393005371,35.268397319259996],
                [-80.70324897766113,35.26503355355979],
                [-80.71088790893555,35.2553619492954],
                [-80.71681022644043,35.2553619492954],
                [-80.7150936126709,35.26054831539319],
                [-80.71869850158691,35.26026797976481],
                [-80.72032928466797,35.26061839914875],
                [-80.72264671325684,35.26033806376283],
                [-80.72487831115721,35.26545403190955]
            ]
        ]
    },
    "id": "102374",
    "properties": {
        "name": "Plaza Road Park"
    }
}`)
}

var complexFeature = `{
  "id": 202418985,
  "type": "Feature",
  "properties": {
    "addr:full":"5607 McKinley Ave Los Angeles CA 90011",
    "addr:housenumber":"5607",
    "addr:postcode":"90011",
    "addr:street":"Mckinley Ave",
    "edtf:cessation":"uuuu",
    "edtf:inception":"uuuu",
    "geom:area":0.0,
    "geom:bbox":"-118.26089,33.99073,-118.26089,33.99073",
    "geom:latitude":33.99073,
    "geom:longitude":-118.26089,
    "iso:country":"US",
    "mz:hierarchy_label":1,
    "sg:address":"5607 McKinley Ave",
    "sg:city":"Los Angeles",
    "sg:classifiers":[
        {
            "category":"Wholesale",
            "subcategory":"Toys & Hobbies",
            "type":"Manufacturing & Wholesale Goods"
        }
    ],
    "sg:owner":"simplegeo",
    "sg:phone":"+1 323 231 0540",
    "sg:postcode":"90011",
    "sg:province":"CA",
    "sg:tags":[
        "wholesaler"
    ],
    "src:geom":"simplegeo",
    "wof:belongsto":[
        85633793,
        85688637
    ],
    "wof:breaches":[],
    "wof:concordances":{
        "sg:id":"SG_0i3ZtGVxBmvnGGcg7wZrlY_33.990730_-118.260890@1294081369"
    },
    "wof:country":"US",
    "wof:geomhash":"fa3426d7a9b6c92b5e2857b4daef560f",
    "wof:hierarchy":[
        {
            "continent_id":-1,
            "country_id":85633793,
            "locality_id":-1,
            "neighbourhood_id":-1,
            "region_id":85688637,
            "venue_id":202418985
        }
    ],
    "wof:id":202418985,
    "wof:lastmodified":1472331065,
    "wof:name":"Hotfix Wholesale Inc",
    "wof:parent_id":-1,
    "wof:placetype":"venue",
    "wof:repo":"whosonfirst-data-venue-us-ca",
    "wof:superseded_by":[],
    "wof:supersedes":[],
    "wof:tags":[
        "wholesaler"
    ],
	"added:by:tidwall": "\n\"\\\\15\u00f8C 3\u0111"
},
  "bbox": [
    -118.26089,
    33.99073,
    -118.26089,
    33.99073
],
  "geometry": {"coordinates":[-118.26089,33.99073],"type":"Point"}
}`

func TestComplexFeature(t *testing.T) {
	testJSON(t, complexFeature)
	o, err := ObjectJSON(complexFeature)
	if err != nil {
		t.Fatal(err)
	}
	o = o
}
