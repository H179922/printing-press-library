// Copyright 2026 H179922 and contributors. Licensed under Apache-2.0. See LICENSE.

package ofa

import (
	"errors"
	"testing"
)

// Synthetic fixtures. The markup mirrors the live backend shape observed
// 2026-10-13 (unclosed result cells, data-appnum/data-chic attributes); the
// dogs, numbers, and registrations are invented. Real OFA results are
// copyrighted and are not committed.
const resultsFixture = `3 matches.  Page 1 of <span id='api_page_count'>1</span><br />
<table id='as_results' class='api_datalist'><thead><tr><th>&nbsp;</th><th>Name</th></tr></thead><tbody>
<tr class='as_results_row' data-appnum='1000001' data-chic='cur' ><td></td>
<td>FAKEKENNEL ALPHA&#039;S STAR</td>
<td>XX00000001</td>
<td>AUSTRALIAN KOOLIE</td>
<td>F</td>
<td> </td>
<td>Dec  1 2023 </td><td>Feb 10 2026 </td>
<td style='text-align:right;'>26</td>
<td>AUK-51G26F-C-VPI</td>
<td>HIPS</td><tbody>
<tr class='as_results_row' data-appnum='1000001' data-chic='cur' ><td></td>
<td>FAKEKENNEL ALPHA&#039;S STAR</td>
<td>XX00000001</td>
<td>AUSTRALIAN KOOLIE</td>
<td>F</td>
<td> </td>
<td>Dec  1 2023 </td><td>Feb 10 2026 </td>
<td style='text-align:right;'>26</td>
<td>AUK-EL51F26-C-VPI</td>
<td>ELBOW</td><tbody>
<tr class='as_results_row' data-appnum='1000002' data-chic='no' ><td></td>
<td>2 BAR MADE UP</td>
<td>XX00000002</td>
<td>AUSTRALIAN KOOLIE</td>
<td>M</td>
<td>RED</td>
<td>Jan 10 2024 </td><td>Jan 14 2026 </td>
<td style='text-align:right;'>24</td>
<td></td>
<td>HIPS</td></tbody></table>`

func TestParseResults(t *testing.T) {
	p, err := ParseResults([]byte(resultsFixture))
	if err != nil {
		t.Fatal(err)
	}
	if p.Total != 3 || p.PageCount != 1 || len(p.Rows) != 3 {
		t.Fatalf("total=%d pages=%d rows=%d, want 3/1/3", p.Total, p.PageCount, len(p.Rows))
	}
	r := p.Rows[0]
	if r.AppNum != "1000001" || !r.CHIC || r.Name != "FAKEKENNEL ALPHA'S STAR" {
		t.Errorf("row0 identity wrong: %+v", r)
	}
	if r.Birthdate != "2023-12-01" || r.TestDate != "2026-02-10" || r.AgeMonths != 26 {
		t.Errorf("row0 dates/age wrong: %+v", r)
	}
	if r.Test != "HIPS" || r.Grade != "GOOD" || r.Flag != "" {
		t.Errorf("row0 hips decode wrong: %+v", r)
	}
	if p.Rows[1].Grade != "" {
		t.Errorf("elbow row must not get a hip grade: %+v", p.Rows[1])
	}
	last := p.Rows[2]
	if last.CHIC || last.Color != "RED" || last.Flag != "no-number" {
		t.Errorf("row2 wrong: %+v", last)
	}
}

func TestParseResultsNoMatchesAndGarbage(t *testing.T) {
	p, err := ParseResults([]byte("No matches found for your search entries.  <br />Please try"))
	if err != nil || p.Total != 0 || len(p.Rows) != 0 {
		t.Errorf("no-match page: %+v %v", p, err)
	}
	_, err = ParseResults([]byte("<html><body>maintenance</body></html>"))
	var ue *UnexpectedResponseError
	if !errors.As(err, &ue) {
		t.Errorf("garbage page should be UnexpectedResponseError, got %v", err)
	}
}

func TestParseResultsRejectsSilentRowLoss(t *testing.T) {
	malformed := "2 matches.  Page 1 of <span id='api_page_count'>1</span>" +
		"<tr class='as_results_row' data-appnum='1' data-chic='no'><td></td><td>ONLY NAME<td>XX1</tbody></table>"
	var ue *UnexpectedResponseError
	if _, err := ParseResults([]byte(malformed)); !errors.As(err, &ue) {
		t.Errorf("row with too few cells must error, got %v", err)
	}
	noRows := "5 matches.  Page 1 of <span id='api_page_count'>1</span><table id='as_results'></table>"
	if _, err := ParseResults([]byte(noRows)); !errors.As(err, &ue) {
		t.Errorf("matches header with zero rows must error, got %v", err)
	}
}

func TestParseResultsThousandsSeparator(t *testing.T) {
	row := "<tr class='as_results_row' data-appnum='1' data-chic='no'><td></td><td>FAKE ONE<td>XX1<td>AUSTRALIAN SHEPHERD<td>M<td> <td>Jan 1 2024<td>Feb 1 2026<td>25<td>AS-1G25M<td>HIPS</tbody></table>"
	p, err := ParseResults([]byte("1,439 matches.  Page 1 of <span id='api_page_count'>72</span>" + row))
	if err != nil || p.Total != 1439 || p.PageCount != 72 {
		t.Errorf("got %+v %v", p, err)
	}
}

const dogFixture = `<input type='hidden' name='api_key' value='1000003'>
<table id='as_detail_individual'>
<tr><td colspan='3'><span id='as_detail_name'>FAKEKENNEL BRAVO</span><td rowspan='6'><div class='as_chic' data-chic='cur'></div>
<!--<tr><td colspan='3'>-->
<tr><td colspan='3'>XX00000003
<tr><td colspan='3'>M BLACK AUSTRALIAN SHEPHERD
<tr><td colspan='3'>Born Apr  5 2021
<tr><td colspan='3'>
<tr><td colspan='3'>Sire: XX00000010<br /> Dam: XX00000011<br />
</table>
<table id='as_detail_tests'>
<thead><tr><th>Registry</th><th>Test Date</th><th>Report Date</th><th>Age</th><th>Conclusion</th><th>OFA Number</th></tr></thead><tbody>
<tr ><td>ELBOW</td><td>Jan 14 2026 </td><td>Jan 29 2026 </td><td style='text-align:right'>24</td><td>NORMAL</td><td>AS-EL50M24-C-NOPI</td></tr>
</tbody><tbody>
<tr ><td>HIPS</td><td>Jan 14 2026 </td><td>Jan 29 2026 </td><td style='text-align:right'>24</td><td>EXCELLENT</td><td>AS-50E24M-C-NOPI</td></tr>
</tbody></table>
<table id='as_detail_relatives'><tbody><tr><th>Full Sibs</th></tr></tbody><tr class='as_relatives_row' data-appnum='1000004' data-chic='no'> <td></td> <td data-sort='X'>FAKEKENNEL CHARLIE</td> <td>F</td> <td>Littermate</td> <td>XX00000004</td> <td>Apr 5 2021 </td> <td><div title='x'>AS-1G24F</div> </tr></table>`

func TestParseDog(t *testing.T) {
	d, err := ParseDog([]byte(dogFixture), "1000003")
	if err != nil {
		t.Fatal(err)
	}
	if d.Name != "FAKEKENNEL BRAVO" || d.Registration != "XX00000003" || !d.CHIC {
		t.Errorf("identity wrong: %+v", d)
	}
	if d.Sex != "M" || d.Color != "BLACK" || d.Breed != "AUSTRALIAN SHEPHERD" || d.Born != "2021-04-05" {
		t.Errorf("sex/color/breed/born wrong: %+v", d)
	}
	if d.Sire != "XX00000010" || d.Dam != "XX00000011" {
		t.Errorf("sire/dam wrong: %q %q", d.Sire, d.Dam)
	}
	if len(d.Tests) != 2 || d.Tests[1].Conclusion != "EXCELLENT" || d.Tests[1].ReportDate != "2026-01-29" {
		t.Errorf("tests wrong: %+v", d.Tests)
	}
	if len(d.Relatives) != 1 || d.Relatives[0].Relation != "Littermate" || d.Relatives[0].AppNum != "1000004" {
		t.Errorf("relatives wrong: %+v", d.Relatives)
	}
}

func TestParseDogNotFound(t *testing.T) {
	if _, err := ParseDog([]byte("<div id='as_detail_full'><table></table></div>"), "1"); !errors.Is(err, ErrDogNotFound) {
		t.Errorf("detail shell without a name: want ErrDogNotFound, got %v", err)
	}
	var ue *UnexpectedResponseError
	for _, body := range []string{"", "query sql: update X set y [Microsoft][ODBC Driver 17 for SQL Server] The statement has been terminated.", "<html>maintenance</html>"} {
		if _, err := ParseDog([]byte(body), "1"); !errors.As(err, &ue) {
			t.Errorf("body %q: want UnexpectedResponseError, got %v", body, err)
		}
	}
}

func TestNormalizeDate(t *testing.T) {
	for in, want := range map[string]string{"Dec  1 2023 ": "2023-12-01", "Feb 10 2026": "2026-02-10", "unknown": "unknown", "": ""} {
		if got := NormalizeDate(in); got != want {
			t.Errorf("NormalizeDate(%q)=%q want %q", in, got, want)
		}
	}
}
