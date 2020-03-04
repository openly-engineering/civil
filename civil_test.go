// Copyright 2016 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package civil implements types for civil time, a time-zone-independent
// representation of time that follows the rules of the proleptic
// Gregorian calendar with exactly 24-hour days, 60-minute hours, and 60-second
// minutes.
//
// Because they lack location information, these types do not represent unique
// moments or intervals of time. Use time.Time for that purpose.
package civil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDate_RoundTrip_JSON(t *testing.T) {
	type TC struct {
		Name            string
		In              Date
		Out             []byte
		HasMarshalErr   bool // tests cases TBD
		HasUnmarshalErr bool
	}
	tcs := []TC{
		TC{Name: "tc0", In: Date{0, 0, 0}, Out: []byte(`"0000-00-00"`)},
		TC{Name: "tc1", In: Date{0, 1, 1}, Out: []byte(`"0000-01-01"`)},
		TC{Name: "tc2", In: Date{2020, 3, 4}, Out: []byte(`"2020-03-04"`)},
		TC{Name: "tc3", In: Date{2345, 12, 25}, Out: []byte(`"2345-12-25"`)},
		/* === ERRORS === */
		TC{Name: "D-bad-day", In: Date{2020, 2, -1}, Out: []byte(`"2020-02--1"`),
			HasUnmarshalErr: true},
		TC{Name: "D-bad-month", In: Date{2020, 13, 1}, Out: []byte(`"2020-13-01"`),
			HasUnmarshalErr: true},
		TC{Name: "D-bad-year", In: Date{-2020, 1, 1}, Out: []byte(`"-2020-01-01"`),
			HasMarshalErr: true},
	}

	for _, tc := range tcs {
		t.Run(tc.Name, func(*testing.T) {
			mBytes, err := tc.In.MarshalJSON()
			switch {
			case tc.HasMarshalErr && err != nil: // expected error
				t.Logf("Got expected error: %v", err)
				return
			case tc.HasMarshalErr && err == nil: // missing error
				t.Fatalf("Missing error, mBytes: %v, string: %s", mBytes, string(mBytes))
			case !tc.HasMarshalErr && err != nil: // unexpected error
				t.Fatalf("Unexpected error: %v", err)
			default:
				t.Logf("no error, mBytes: %v, string: %s", mBytes, string(mBytes))
				assert.Equal(t, tc.Out, mBytes)
			}

			if t.Failed() {
				return
			}

			if string(mBytes) != string(tc.Out) {
				t.Errorf("!= bytes :: json: %v, Out: %v", mBytes, tc.Out)
				t.Errorf("!= string:: json: %s, Out: %s", string(mBytes), string(tc.Out))
				t.Fatalf("!= In: %d", tc.In)
			}
			t.Logf("tc.In: %v, mBytes: %s, len(mBytes): %d, err: %v", tc.In, string(mBytes), len(mBytes), err)

			// Now Unmarshal to confirm we get what we had
			var d0 Date

			err = d0.UnmarshalJSON(mBytes)
			t.Logf("d0: %v, mBytes: %s, unmarshal err: %v", d0, string(mBytes), err)

			switch {
			case tc.HasUnmarshalErr && err != nil: // expected error
				t.Logf("Got expected error: %v", err)
				return
			case tc.HasUnmarshalErr && err == nil: // missing error
				t.Fatalf("Missing error, mBytes: %v, string: %s", mBytes, string(mBytes))
			case !tc.HasUnmarshalErr && err != nil: // unexpected error
				t.Fatalf("Unexpected error: %v", err)
			default:
				assert.Equal(t, tc.Out, mBytes)
			}
		})

	}
}

func TestDate_MarshalJSON(t *testing.T) {

	dLeap := Date{
		Year:  2020,
		Month: 2,
		Day:   29,
	}

	json, err := dLeap.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, []byte(`"2020-02-29"`), json)

	dInvalid := Date{
		Year:  -1,
		Month: 2,
		Day:   29,
	}

	json, err = dInvalid.MarshalJSON()
	assert.EqualError(t, err, "Date.MarshalJSON: year '-1' outside of range [0,9999]")
	assert.Nil(t, json)
}

func TestDate_UnmarshalJSON(t *testing.T) {
	jsonLeap := []byte(`"2020-02-29"`)
	dLeap := &Date{}
	err := dLeap.UnmarshalJSON(jsonLeap)
	assert.NoError(t, err)
	assert.Equal(t, Date{Year: 2020, Month: 2, Day: 29}, *dLeap)

	jsonInvalid := []byte(`"2020-13-40"`)
	dInvalid := &Date{}
	err = dInvalid.UnmarshalJSON(jsonInvalid)
	assert.NotNil(t, err)
}

func TestDate_AddMonths(t *testing.T) {
	dLeap := Date{
		Year:  2020,
		Month: 2,
		Day:   29,
	}

	dLeap = dLeap.AddMonths(12)

	assert.Equal(t, Date{Year: 2021, Month: 3, Day: 1}, dLeap) // no leap day in 2021, so pushes over to 3/1
}

func TestDate_AddYears(t *testing.T) {
	dLeap := Date{
		Year:  2020,
		Month: 2,
		Day:   29,
	}

	dLeap = dLeap.AddYears(1)

	assert.Equal(t, Date{Year: 2021, Month: 3, Day: 1}, dLeap) // no leap day in 2021, so pushes over to 3/1
}

func TestDate_Value(t *testing.T) {
	d := Date{
		Year:  2020,
		Month: 2,
		Day:   29,
	}

	v, err := d.Value()
	assert.NoError(t, err)
	assert.Equal(t, v, "2020-02-29")

	d1 := Date{Year: 0, Month: 0, Day: 0}
	v1, err1 := d1.Value()
	assert.NoError(t, err1)
	assert.Equal(t, v1, "0000-00-00")
}

func TestDate_Scan_String(t *testing.T) {
	d := &Date{}
	var v interface{}
	v = "2020-02-29"
	d.Scan(v)
	assert.Equal(t, Date{Year: 2020, Month: 2, Day: 29}, *d)
}

func TestDate_Scan_Time(t *testing.T) {
	d := &Date{}
	var v interface{}
	v = time.Date(2020, time.February, 29, 0, 0, 0, 0, time.UTC)
	d.Scan(v)
	assert.Equal(t, Date{Year: 2020, Month: 2, Day: 29}, *d)
}

func TestTime_MarshalJSON(t *testing.T) {
	time := Time{
		Hour:       3,
		Minute:     42,
		Second:     31,
		Nanosecond: 876,
	}

	json, err := time.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, []byte(`"03:42:31.000000876"`), json)
}

func TestTime0_MarshalJSON(t *testing.T) {
	time := Time{
		Hour:       0,
		Minute:     0,
		Second:     0,
		Nanosecond: 0,
	}

	json, err := time.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, []byte(`"00:00:00"`), json)
}

func TestTime_UnmarshalJSON(t *testing.T) {
	jsonGood := []byte(`"03:42:31.000000876"`)
	timeGood := &Time{}
	err := timeGood.UnmarshalJSON(jsonGood)
	assert.NoError(t, err)
	assert.Equal(t, Time{Hour: 3, Minute: 42, Second: 31, Nanosecond: 876}, *timeGood)

	jsonInvalid := []byte(`"-3:42:31.000000876"`)
	timeInvalid := &Time{}
	err = timeInvalid.UnmarshalJSON(jsonInvalid)
	assert.EqualError(t, err, "invalid time: parsing time \"-3:42:31.000000876\" as \"15:04:05.999999999\": cannot parse \"-3:42:31.000000876\" as \"15\"")
}

func TestTime_Value(t *testing.T) {
	time := Time{
		Hour:       3,
		Minute:     42,
		Second:     31,
		Nanosecond: 876,
	}

	v, err := time.Value()
	assert.NoError(t, err)
	assert.Equal(t, "03:42:31.000000876", v)
}

func TestTime_Scan_String(t *testing.T) {
	time := &Time{}
	var v interface{}
	v = "03:42:31.000000876"
	time.Scan(v)
	assert.Equal(t, *time, Time{Hour: 3, Minute: 42, Second: 31, Nanosecond: 876})
}

func TestTime_Scan_Time(t *testing.T) {
	tm := &Time{}
	var v interface{}
	v = time.Date(2020, time.February, 29, 3, 42, 31, 876, time.UTC)
	tm.Scan(v)
	assert.Equal(t, *tm, Time{Hour: 3, Minute: 42, Second: 31, Nanosecond: 876})
}

func TestDateTime_MarshalJSON(t *testing.T) {

	datetime := DateTime{
		Date: Date{
			Year:  2020,
			Month: 2,
			Day:   29,
		},
		Time: Time{
			Hour:       3,
			Minute:     42,
			Second:     31,
			Nanosecond: 876,
		},
	}

	json, err := datetime.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, []byte(`"2020-02-29T03:42:31.000000876"`), json)
}

func TestDateTime_RoundTrip_JSON(t *testing.T) {
	type TC struct {
		Name         string
		In           DateTime
		Out          []byte
		MarshalErr   bool
		UnmarshalErr bool
	}

	tcs := []TC{
		TC{Name: "DT-zero_date-ok_time", In: DateTime{Date{0, 0, 00}, Time{3, 42, 31, 876}},
			Out: []byte(`"0000-00-00T03:42:31.000000876"`)},
		TC{Name: "DT-zero_date-zero_time", In: DateTime{Date{0, 0, 00}, Time{0, 0, 0, 0}},
			Out: []byte(`"0000-00-00T00:00:00"`)},
		TC{Name: "DT-go-zero", In: DateTime{},
			Out: []byte(`"0000-00-00T00:00:00"`)},
		TC{Name: "DT-leap-day", In: DateTime{Date{2020, 2, 29}, Time{3, 42, 31, 876}},
			Out: []byte(`"2020-02-29T03:42:31.000000876"`)},
		TC{Name: "DT-prior-xmas-day", In: DateTime{Date{2019, 12, 25}, Time{1, 2, 3, 4}},
			Out: []byte(`"2019-12-25T01:02:03.000000004"`)},
		TC{Name: "DT-future-xmas-day", In: DateTime{Date{2345, 12, 25}, Time{12, 23, 34, 45}},
			Out: []byte(`"2345-12-25T12:23:34.000000045"`)},
		TC{Name: "DT-first-day", In: DateTime{Date{0, 1, 1}, Time{}},
			Out: []byte(`"0000-01-01T00:00:00"`)},
		TC{Name: "DT-last-day", In: DateTime{Date{9999, 12, 31}, Time{23, 59, 59, 999999999}},
			Out: []byte(`"9999-12-31T23:59:59.999999999"`)},
		/* === ERRORS === */
		TC{Name: "DT-bad-day", In: DateTime{Date{2020, 3, -4}, Time{12, 23, 34, 5}},
			Out:          []byte(`"2020-03--4T12:23:34.000000005"`),
			UnmarshalErr: true},
		TC{Name: "DT-bad-month", In: DateTime{Date{2020, 13, 4}, Time{12, 23, 34, 5}},
			Out:          []byte(`"2020-13-04T12:23:34.000000005"`),
			UnmarshalErr: true},
		TC{Name: "DT-bad-year", In: DateTime{Date{-2020, 3, 4}, Time{12, 23, 34, 5}},
			Out:          []byte(`"-2020-03-04T12:23:34.000000005"`),
			UnmarshalErr: true},
		TC{Name: "DT-bad-hour", In: DateTime{Date{2020, 3, 4}, Time{24, 0, 0, 0}},
			Out:          []byte(`"2020-03-04T24:00:00"`),
			UnmarshalErr: true},
		TC{Name: "DT-bad-minute", In: DateTime{Date{2020, 3, 4}, Time{0, -1, 0, 0}},
			Out:          []byte(`"2020-03-04T00:-1:00"`),
			UnmarshalErr: true},
		TC{Name: "DT-bad-second", In: DateTime{Date{2020, 3, 4}, Time{1, 0, 75, 0}},
			Out:          []byte(`"2020-03-04T01:00:75"`),
			UnmarshalErr: true},
		TC{Name: "DT-bad-ns", In: DateTime{Date{2020, 3, 4}, Time{12, 23, 34, 1231231234}},
			Out:          []byte(`"2020-03-04T12:23:34.1231231234"`),
			UnmarshalErr: true},
	}

	for _, tc := range tcs {
		t.Run(tc.Name, func(t *testing.T) {
			dtBytes, err := tc.In.MarshalJSON()

			switch {
			case tc.MarshalErr && err != nil: // exected and got an error
				t.Logf("Got expected MarshalJSON error: %v", err)
				return
			case tc.MarshalErr && err == nil: // missing error
				t.Fatalf("Did not get expected error, dtBytes: %v, %s", dtBytes, string(dtBytes))
			case !tc.MarshalErr && err != nil: // unexpected error
				t.Fatalf("Got UNexpected error: %v", err)
			default: // all good
				t.Logf("No MarshalJSON error, dtBytes: %s", string(dtBytes))
				assert.NoError(t, err)
				assert.Equal(t, tc.Out, dtBytes)
			}

			var dt DateTime
			err = dt.UnmarshalJSON(dtBytes)

			switch {
			case tc.UnmarshalErr && err != nil: // expected and got an error
				t.Logf("Got expected UnmarshalJSON error: %v", err)
				return
			case tc.UnmarshalErr && err == nil: // missing error
				t.Fatalf("Did not get expected error, dt: %v", dt)
			case !tc.UnmarshalErr && err != nil: // Unexpected error
				t.Fatalf("Got UNexpected error: %v", err)
			default:
				t.Logf("No UnmarshalJSON error, dt: %v", dt)
				assert.NoError(t, err)
				assert.Equal(t, tc.In, dt)
			}
		})
	}
}

func TestDateTime_UnmarshalJSON(t *testing.T) {
	jsonGood := []byte(`"2020-02-29T03:42:31.000000876"`)
	datetimeGood := &DateTime{}
	err := datetimeGood.UnmarshalJSON(jsonGood)
	assert.NoError(t, err)
	expectedGood := DateTime{
		Date: Date{
			Year:  2020,
			Month: 2,
			Day:   29,
		},
		Time: Time{
			Hour:       3,
			Minute:     42,
			Second:     31,
			Nanosecond: 876,
		},
	}
	assert.Equal(t, expectedGood, *datetimeGood)

	jsonInvalid := []byte(`"0-02-29T03:42:31.000000876"`)
	datetimeInvalid := &DateTime{}
	err = datetimeInvalid.UnmarshalJSON(jsonInvalid)
	assert.NotNil(t, err)
}

func TestDateTime_Value(t *testing.T) {
	datetime := DateTime{
		Date: Date{
			Year:  2020,
			Month: 2,
			Day:   29,
		},
		Time: Time{
			Hour:       3,
			Minute:     42,
			Second:     31,
			Nanosecond: 876,
		},
	}

	v, err := datetime.Value()
	assert.NoError(t, err)
	assert.Equal(t, "2020-02-29T03:42:31.000000876", v)
}

func TestDateTime_ScanString(t *testing.T) {
	datetime := &DateTime{}
	var v interface{}
	v = "2020-02-29T03:42:31.000000876"
	datetime.Scan(v)
	expected := DateTime{
		Date: Date{
			Year:  2020,
			Month: 2,
			Day:   29,
		},
		Time: Time{
			Hour:       3,
			Minute:     42,
			Second:     31,
			Nanosecond: 876,
		},
	}
	assert.Equal(t, *datetime, expected)
}

func TestDateTime_Scan(t *testing.T) {
	datetime := &DateTime{}
	var v interface{}
	v = time.Date(2020, time.February, 29, 3, 42, 31, 876, time.UTC)
	datetime.Scan(v)
	expected := DateTime{
		Date: Date{
			Year:  2020,
			Month: 2,
			Day:   29,
		},
		Time: Time{
			Hour:       3,
			Minute:     42,
			Second:     31,
			Nanosecond: 876,
		},
	}
	assert.Equal(t, *datetime, expected)
}
