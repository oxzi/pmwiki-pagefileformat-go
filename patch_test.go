package pmwiki

import (
	"strings"
	"testing"
)

func TestPatchApplyWikipediaExample(t *testing.T) {
	// Example from <https://en.wikipedia.org/wiki/Diff#Usage>

	input := "This part of the\ndocument has stayed the\nsame from version to\nversion.  It shouldn't\n" +
		"be shown if it doesn't\nchange.  Otherwise, that\nwould not be helping to\ncompress the size of the\n" +
		"changes.\n\nThis paragraph contains\ntext that is outdated.\nIt will be deleted in the\n" +
		"near future.\n\nIt is important to spell\ncheck this dokument. On\nthe other hand, a\n" +
		"misspelled word isn't\nthe end of the world.\nNothing in the rest of\nthis paragraph needs to\n" +
		"be changed. Things can\nbe added after it."
	output := "This is an important\nnotice! It should\ntherefore be located at\nthe beginning of this\n" +
		"document!\n\nThis part of the\ndocument has stayed the\nsame from version to\nversion.  It shouldn't\n" + "" +
		"be shown if it doesn't\nchange.  Otherwise, that\nwould not be helping to\ncompress the size of the\n" +
		"changes.\n\nIt is important to spell\ncheck this document. On\nthe other hand, a\nmisspelled word isn't\n" +
		"the end of the world.\nNothing in the rest of\nthis paragraph needs to\nbe changed. Things can\n" +
		"be added after it.\n\nThis paragraph contains\nimportant new additions\nto this document.\n"
	diff := "0a1,6\n> This is an important\n> notice! It should\n> therefore be located at\n" +
		"> the beginning of this\n> document!\n>\n11,15d16\n< This paragraph contains\n" +
		"< text that is outdated.\n< It will be deleted in the\n< near future.\n<\n17c18\n" +
		"< check this dokument. On\n---\n> check this document. On\n24a26,29\n>\n" +
		"> This paragraph contains\n> important new additions\n> to this document.\n"

	patch, err := parsePatch(diff)
	if err != nil {
		t.Fatal(err)
	}

	var outputBuilder strings.Builder
	if err := patch.Apply(strings.NewReader(input), &outputBuilder); err != nil {
		t.Fatal(err)
	} else if outputRes := outputBuilder.String(); outputRes != output {
		t.Fatal("output differs")
	}
}

func TestPatchApplyPmWiki(t *testing.T) {
	// Page from hsmr wiki <https://hsmr.cc/Events/2020-11-07-Siebter>

	input := "version=pmwiki-2.2.106 ordered=1 urlencoded=1\nauthor=oxzi\nhost=fe80::1\nname=Events.2020-11-07-Siebter\nrev=4\n" +
		"text=(:if false:)%0a(:title Siebter des Monats:) %0aStartYear: 2020%0aStartMonth: 11%0aStartDay: 07%0aStartTime: 20:00%0aEventLocation: Internet%0a%0a---- The following fields can be omitted: %0aEndTime: %0aEndDay: %0aEndMonth: %0aEndYear: %0aEventDescription: Digitales Beisammensein; guter Termin zum (wieder) {-reinschauen-}{+kennenlernen+}!%0a(:ifend:)%0a''2nd wave is happening-Edition''%0a%0aDieses Mal online! Mit [[Hackslam/2020-11-07 | Hackslam]].%0a%0a[@%0a/***%0a *     _%0a *    | |_ ___ _____ ___%0a *    |   |_ -|     |  _|%0a *    |_|_|___|_|_|_|_|%0a *%0a *%0a * Diese Einladung ist Teil vom Siebten des Monats.%0a *%0a * Der Siebter des Monats ist einer freie Veranstaltung: Sie können%0a * diese unter den Bedingungen der Umgangsvereinbarung, wie von dem%0a * [hsmr], Version 2017-09-03 oder (nach Ihrer Wahl) jeder neueren%0a * veröffentlichten Version, weiter beiwohnen und/oder modifizieren.%0a *%0a * Der Siebte des Monats wird in der Hoffnung, dass er nützlich sein%0a * wird, aber OHNE JEDE GEWÄHRLEISTUNG, bereitgestellt; sogar ohne die%0a * implizite Gewährleistung der MARKTFÄHIGKEIT oder des STATTFINDENS.%0a *%0a * Sie sollten eine Kopie der Umgangsvereinbarung zusammen mit dieser%0a * Einladung erhalten haben. Wenn nicht, siehe%0a * %3chttps://hsmr.cc/Main/Umgangsvereinbarung>.%0a */%0a%0a#include %3cstdio.h>%0a#include %3cstdlib.h>%0a%0a#define EVENT_NAME \"Siebter des Monats / Hackslam\"%0a#define EVENT_DATE \"2020-11-07, 20:00 Uhr\"%0a%0aint%0amain (int argc, char **argv)%0a{%0a  printf (\"\\%0aBenutzung: %25s\\n\\%0a\\n\\%0aZum kommenden %25s am %25s lädt\\n\\%0ader Marburger Hackspace [hsmr] zu einem offenen Abend mit\\n\\%0aVortraegen und gemeinsamen VoIPen ein.\\n\\%0a\\n\\%0a  * https://hsmr.cc/\\n\\%0a  * https://hsmr.cc/Hackslam/2020-11-07\\n\\%0a  * mailto:public@lists.hsmr.cc\\n\\%0a  * ircs://irc.hackint.org/hsmr\\n\\%0a\\n\\%0aAktuell suchen wir noch Programmpunkte. Falls Du also Lust hast etwas\\n\\%0azwischen Haiku und Death by PowerPoint\\u2122 vorzutragen, dann fühle\\n\\%0adich herzlich eingeladen.\\n\", EVENT_NAME, EVENT_NAME, EVENT_DATE);%0a%0a  return EXIT_SUCCESS;%0a}%0a@]\n" + "" +
		"time=1603451578\ntitle=Siebter des Monats\n" + "" +
		"author:1603451578=oxzi\ndiff:1603451578:1603450377:=6c6%0a%3c StartTime: 20:00%0a---%0a> StartTime: 18:00%0a\nhost:1603451578=fe80::1\n" +
		"author:1603450377=oxzi\ncsum:1603450377=applied \"[PATCH] fix: double of 'Uhr'\"\ndiff:1603450377:1603446769:=57c57%0a%3c der Marburger Hackspace [hsmr] zu einem offenen Abend mit\\n\\%0a---%0a> der Marburger Hackspace [hsmr] Uhr zu einem offenen Abend mit\\n\\%0a\nhost:1603450377=fe80::1\n" + "" +
		"author:1603446769=oxzi\ndiff:1603446769:1603390737:=19,71c19,27%0a%3c %0a%3c [@%0a%3c /***%0a%3c  *     _%0a%3c  *    | |_ ___ _____ ___%0a%3c  *    |   |_ -|     |  _|%0a%3c  *    |_|_|___|_|_|_|_|%0a%3c  *%0a%3c  *%0a%3c  * Diese Einladung ist Teil vom Siebten des Monats.%0a%3c  *%0a%3c  * Der Siebter des Monats ist einer freie Veranstaltung: Sie können%0a%3c  * diese unter den Bedingungen der Umgangsvereinbarung, wie von dem%0a%3c  * [hsmr], Version 2017-09-03 oder (nach Ihrer Wahl) jeder neueren%0a%3c  * veröffentlichten Version, weiter beiwohnen und/oder modifizieren.%0a%3c  *%0a%3c  * Der Siebte des Monats wird in der Hoffnung, dass er nützlich sein%0a%3c  * wird, aber OHNE JEDE GEWÄHRLEISTUNG, bereitgestellt; sogar ohne die%0a%3c  * implizite Gewährleistung der MARKTFÄHIGKEIT oder des STATTFINDENS.%0a%3c  *%0a%3c  * Sie sollten eine Kopie der Umgangsvereinbarung zusammen mit dieser%0a%3c  * Einladung erhalten haben. Wenn nicht, siehe%0a%3c  * %3chttps://hsmr.cc/Main/Umgangsvereinbarung>.%0a%3c  */%0a%3c %0a%3c #include %3cstdio.h>%0a%3c #include %3cstdlib.h>%0a%3c %0a%3c #define EVENT_NAME \"Siebter des Monats / Hackslam\"%0a%3c #define EVENT_DATE \"2020-11-07, 20:00 Uhr\"%0a%3c %0a%3c int%0a%3c main (int argc, char **argv)%0a%3c {%0a%3c   printf (\"\\%0a%3c Benutzung: %25s\\n\\%0a%3c \\n\\%0a%3c Zum kommenden %25s am %25s lädt\\n\\%0a%3c der Marburger Hackspace [hsmr] Uhr zu einem offenen Abend mit\\n\\%0a%3c Vortraegen und gemeinsamen VoIPen ein.\\n\\%0a%3c \\n\\%0a%3c   * https://hsmr.cc/\\n\\%0a%3c   * https://hsmr.cc/Hackslam/2020-11-07\\n\\%0a%3c   * mailto:public@lists.hsmr.cc\\n\\%0a%3c   * ircs://irc.hackint.org/hsmr\\n\\%0a%3c \\n\\%0a%3c Aktuell suchen wir noch Programmpunkte. Falls Du also Lust hast etwas\\n\\%0a%3c zwischen Haiku und Death by PowerPoint\\u2122 vorzutragen, dann fühle\\n\\%0a%3c dich herzlich eingeladen.\\n\", EVENT_NAME, EVENT_NAME, EVENT_DATE);%0a%3c %0a%3c   return EXIT_SUCCESS;%0a%3c }%0a%3c @]%0a\\ No newline at end of file%0a---%0a> (:if false:)%0a> Notes: %0a> * Day, Month, Year: e.g. 15, 07, 2013 (two/four digits format required for proper sorting)%0a> * Time: e.g. 14:30, 09:00 (use 24 hours format and leading zeros for proper sorting)%0a> * Please leave the \"if false\" and \"ifend\" fields untouched - they are needed by PmWiki%0a> * Please fill in the :title : field like this: \":title Summer Party:\" with no colon in between (and please don't remove the brackets)%0a> * If the title is empty the pagename will become the title%0a> * Additional content (text, images, attachments, links...) can be added after the \"ifend\" tags%0a> (:ifend:)%0a\\ No newline at end of file%0a\nhost:1603446769=fe80::1\n" +
		"author:1603390737=xkey\ndiff:1603390737:1603390737:=1,27d0%0a%3c (:if false:)%0a%3c (:title Siebter des Monats:) %0a%3c StartYear: 2020%0a%3c StartMonth: 11%0a%3c StartDay: 07%0a%3c StartTime: 18:00%0a%3c EventLocation: Internet%0a%3c %0a%3c ---- The following fields can be omitted: %0a%3c EndTime: %0a%3c EndDay: %0a%3c EndMonth: %0a%3c EndYear: %0a%3c EventDescription: Digitales Beisammensein; guter Termin zum (wieder) {-reinschauen-}{+kennenlernen+}!%0a%3c (:ifend:)%0a%3c ''2nd wave is happening-Edition''%0a%3c %0a%3c Dieses Mal online! Mit [[Hackslam/2020-11-07 | Hackslam]].%0a%3c (:if false:)%0a%3c Notes: %0a%3c * Day, Month, Year: e.g. 15, 07, 2013 (two/four digits format required for proper sorting)%0a%3c * Time: e.g. 14:30, 09:00 (use 24 hours format and leading zeros for proper sorting)%0a%3c * Please leave the \"if false\" and \"ifend\" fields untouched - they are needed by PmWiki%0a%3c * Please fill in the :title : field like this: \":title Summer Party:\" with no colon in between (and please don't remove the brackets)%0a%3c * If the title is empty the pagename will become the title%0a%3c * Additional content (text, images, attachments, links...) can be added after the \"ifend\" tags%0a%3c (:ifend:)%0a\\ No newline at end of file%0a\nhost:1603390737=fe80::2\n"

	pageFile, err := ParsePageFile(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	text := pageFile.Text
	for rev := pageFile.Revs[pageFile.Time]; text != ""; rev = pageFile.Revs[rev.DiffAgainst] {
		var textOut strings.Builder
		if err := rev.Diff.Apply(strings.NewReader(text), &textOut); err != nil {
			t.Fatal(err)
		}
		text = textOut.String()
	}
}
