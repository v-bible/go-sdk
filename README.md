<div align="center">

  <h1>go-sdk</h1>

  <p>
    Golang SDK for v-bible
  </p>

<!-- Badges -->
<p>
  <a href="https://github.com/v-bible/go-sdk/graphs/contributors">
    <img src="https://img.shields.io/github/contributors/v-bible/go-sdk" alt="contributors" />
  </a>
  <a href="">
    <img src="https://img.shields.io/github/last-commit/v-bible/go-sdk" alt="last update" />
  </a>
  <a href="https://github.com/v-bible/go-sdk/network/members">
    <img src="https://img.shields.io/github/forks/v-bible/go-sdk" alt="forks" />
  </a>
  <a href="https://github.com/v-bible/go-sdk/stargazers">
    <img src="https://img.shields.io/github/stars/v-bible/go-sdk" alt="stars" />
  </a>
  <a href="https://github.com/v-bible/go-sdk/issues/">
    <img src="https://img.shields.io/github/issues/v-bible/go-sdk" alt="open issues" />
  </a>
  <a href="https://github.com/v-bible/go-sdk/blob/main/LICENSE.md">
    <img src="https://img.shields.io/github/license/v-bible/go-sdk.svg" alt="license" />
  </a>
</p>

<h4>
    <a href="https://github.com/v-bible/go-sdk/">View Demo</a>
  <span> · </span>
    <a href="https://github.com/v-bible/go-sdk">Documentation</a>
  <span> · </span>
    <a href="https://github.com/v-bible/go-sdk/issues/">Report Bug</a>
  <span> · </span>
    <a href="https://github.com/v-bible/go-sdk/issues/">Request Feature</a>
  </h4>
</div>

<br />

<!-- Table of Contents -->

# :notebook_with_decorative_cover: Table of Contents

- [Getting Started](#toolbox-getting-started)
  - [Prerequisites](#bangbang-prerequisites)
  - [Run Locally](#running-run-locally)
- [Usage](#eyes-usage)
  - [Utils](#utils)
    - [Generate Liturgical Calendar](#generate-liturgical-calendar)
    - [Process Verse](#process-verse)
    - [Verse Parse](#verse-parse)
  - [Scripts](#scripts)
- [Roadmap](#compass-roadmap)
- [Contributing](#wave-contributing)
  - [Code of Conduct](#scroll-code-of-conduct)
- [License](#warning-license)
- [Contact](#handshake-contact)
- [Acknowledgements](#gem-acknowledgements)

<!-- Getting Started -->

## :toolbox: Getting Started

<!-- Prerequisites -->

### :bangbang: Prerequisites

- Go: `1.22.3`.

- Go tools:

  All required Go tools is placed in `internal/tools/tools.go`:

  ```bash
  make download-deps
  ```

> [!NOTE]
> These dependencies are not included during build.

<!-- Run Locally -->

### :running: Run Locally

Clone the project:

```bash
git clone https://github.com/v-bible/go-sdk.git
```

Go to the project directory:

```bash
cd go-sdk
```

Install dependencies:

```bash
go mod download
```

<!-- Usage -->

## :eyes: Usage

### Utils

#### Generate Liturgical Calendar

The liturgical data is collected from [The Lectionary for Mass (1998/2002 USA
Edition)](https://catholic-resources.org/Lectionary/1998USL.htm), which is
compiled by Felix Just, S.J., Ph.D. The data is stored in
[v-bible/static](https://github.com/v-bible/static) repository.

Some considerations when generating the liturgical calendar:

- Currently, I don't have "The Lectionary for Mass" book to verify the data. If
  you find any mistakes, please report them to me.

- The verse for the liturgical may varies from different languages and
  translations. Compare the liturgical for the same day **04/03/2024** (Monday of the Third Week of Lent) from
  [vaticanews.va](https://vaticannews.va/):
  in French, Español,
  Vietnamese, and English:

  - [French](https://www.vaticannews.va/fr/evangile-du-jour.html):
    - First Reading: `2 R 5,1-15a`.
    - Gospel: `Lc 4,24-30`.
  - [Español](https://www.vaticannews.va/es/evangelio-de-hoy.html):
    - First Reading: `2 Reyes 5,1-15`.
    - Gospel: `Lc 4,24-30`.
  - [Vietnamese](https://www.vaticannews.va/vi/loi-chua-hang-ngay.html):
    - First Reading: `2 V 5,1-15a`.
    - Gospel: `Lc 4,24-3`.
  - [English](https://www.vaticannews.va/en/word-of-the-day.html):
    - First Reading: `2 Kgs 5:1-15ab`.
    - Gospel: `Lk 4:24-30`.
  - [v-bible/static](https://github.com/v-bible/static):
    - First Reading: `2 Kgs 5:1-15a`.
    - Gospel: `Luke 4:24-30`.

- In the same day may have multiple data for the liturgical (additional
  celebrations, feasts or solemnities may vary from different countries).

- The liturgical calendar also changes based on options:

  - Is Epiphany on 6th January or Sunday after 1st January?
  - Is Ascension on Thursday or Sunday after 40 days of Easter?
  - Special celebrations for each country.

- User can also provide user-defined data for the liturgical calendar.

### Process Verse

Convert verse data to markdown and HTML format.

> [!NOTE]
> For markdown format, your markdown processor SHOULD support [GFM
> footnotes](https://github.blog/changelog/2021-09-30-footnotes-now-supported-in-markdown-fields/).

### Verse Parse

This util comply with the
[Biblical References: Format, Examples,
History](https://catholic-resources.org/Bible/Biblical_References.htm) format.

You can use function `NormalizeQueryUs` or `NormalizeQueryEu` to normalize query
to smaller chapters, ensuring each chapters will have start and end verse. The
`*` means taking all verses. For example:

<!-- prettier-ignore-start -->

| **American Abbr.**                      | **NormalizeQueryUs**                   | **Biblical Passage**                                                                                      | **European Abbr.**      | **NormalizeQueryEu**                   |
|-----------------------------------------|----------------------------------------|-----------------------------------------------------------------------------------------------------------|-------------------------|----------------------------------------|
| John 9                                  | 9:\*-\*                                | The Gospel according to John, chapter 9                                                                   | John 9                  | 9,\*-\*                                |
| John 9, 12                              | 9:\*-\*;12:\*-\*                       | John, chapters 9 and 12 (two chapters only)                                                               | John 9; 12              | 9,\*-\*;12,\*-\*                       |
| John 9--12                              | 9:\*-\*;10:\*-\*;11:\*-\*;12:\*-\*     | John, chapters 9 through 12 (four chaps. total)                                                           | John 9--12              | 9,\*-\*;10,\*-\*;11,\*-\*;12,\*-\*     |
| John 9:12                               | 9:12-12                                | John, chapter 9, verse 12 (only one verse)                                                                | John 9,12               | 9,12-12                                |
| John 9:12b                              | 9:12b-12b                              | John, chapter 9, only the second part of verse 12                                                         | John 9,12b              | 9,12b-12b                              |
| John 9:1, 12                            | 9:1-1;9:12-12                          | John, chapter 9, verses 1 and 12 only                                                                     | John 9,1.12             | 9,1-1;9,12-12                          |
| John 9:1-12                             | 9:1-12                                 | John, chapter 9, the passage from verse 1 to verse 12                                                     | John 9,1-12             | 9,1-12                                 |
| John 9:1-12, 36                         | 9:1-12;9:36-36                         | John, chapter 9, from verse 1 to verse 12, and verse 36                                                   | John 9,1-12.36          | 9,1-12;9,36-36                         |
| John 9:1; 12:36                         | 9:1-1;12:36-36                         | John, only the two verses 9:1 and 12:36                                                                   | John 9,1; 12,36         | 9,1-1;12,36-36                         |
| John 9:1--12:36                         | 9:1-\*;10:\*-\*;11:\*-\*;12:\*-36      | John, the whole section from 9:1 to 12:36                                                                 | John 9,1--12,36         | 9,1-\*;10,\*-\*;11,\*-\*;12,\*-36      |
| John 9:1-12; 12:3-6                     | 9:1-12;12:3-6                          | John, the two passages 9:1-12 and 12:3-6                                                                  | John 9,1-12; 12,3-6     | 9,1-12;12,3-6                          |
| John 9:1-3, 6-12; 12:3-6                | 9:1-3;9:6-12;12:3-6                    | three passages: John 9:1-3; and 9:6-12; and John 12:3-6                                                   | John 9,1-3.6-12; 12,3-6 | 9,1-3;9,6-12;12,3-6                    |
| John 9:1-3, 6-12--12:3-6 (Additional)   | 9:1-3;9:6-\*;10:\*-\*;11:\*-\*;12:\*-6 | John, passage 9:1-3, whole section from 9:6 to 12:6                                                       | John 9,1-3.6-12--12,3-6 | 9,1-3;9,6-\*;10,\*-\*;11,\*-\*;12,\*-6 |
| John 9:12-13                            | 9:12-13                                | John, chapter 9, verses 12 and 13 ("12 and following")                                                    | John 9,12f              | **Not supported**                      |
| not used; better to list exact verse #s | **Not supported**                      | John, chapter 9, verse 12 "and the following verses"; but how many? the end of the text is not specified! | John 9,12ff             | **Not supported**                      |

<!-- prettier-ignore-end -->

> [!NOTE]
> For chapter range, you MUST use `--` (two hyphens) not special characters, the
> same for verse range `-`.

For parsing verses:

- Each verses will have `Number` and `Order`, with `Order` starts from `0` for
  `a`, `1` for `b`, `2` for `c`, etc.
- `*` is converted to `Number`: `-1` and `Order` :`-1`.
- `9:12b` is converted `Number`: `12` and `Order` :`1`.

## Scripts

- `scripts/gen-calendar.go`: Generate liturgical calendar.

<!-- Roadmap -->

## :compass: Roadmap

- [ ] ktcgkpv: Add footnotes for proper names.
- [ ] Add tests.

<!-- Contributing -->

## :wave: Contributing

<a href="https://github.com/v-bible/go-sdk/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=v-bible/go-sdk" />
</a>

Contributions are always welcome!

Please read the [contribution guidelines](./CONTRIBUTING.md).

<!-- Code of Conduct -->

### :scroll: Code of Conduct

Please read the [Code of Conduct](./CODE_OF_CONDUCT.md).

<!-- License -->

## :warning: License

This project is licensed under the **Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International (CC BY-NC-SA 4.0)** License.

[![License: CC BY-NC-SA 4.0](https://licensebuttons.net/l/by-nc-sa/4.0/88x31.png)](https://creativecommons.org/licenses/by-nc-sa/4.0/).

See the **[LICENSE.md](./LICENSE.md)** file for full details.

<!-- Contact -->

## :handshake: Contact

Duong Vinh - tienvinh.duong4@gmail.com

Project Link: [https://github.com/v-bible/go-sdk](https://github.com/v-bible/go-sdk).

<!-- Acknowledgments -->

## :gem: Acknowledgements

Here are useful resources and libraries that we have used in our projects:

- [The Lectionary for Mass (1998/2002 USA
  Edition)](https://catholic-resources.org/Lectionary/1998USL.htm): compiled by
  Felix Just, S.J., Ph.D.
- [Electronic New Testament Educational
  Resources](https://catholic-resources.org/Bible/index.html): compiled by
  Felix Just, S.J., Ph.D.
- [Biblical Book Names &
  Abbreviations](https://catholic-resources.org/Bible/Abbreviations-Abreviaciones.htm):
  compiled by Felix Just, S.J., Ph.D.
- [Calendar of Lectionary Cycles and Movable Liturgical Feasts (1969 – 2100)](https://catholic-resources.org/Lectionary/Calendar.htm): compiled by
  Felix Just, S.J., Ph.D.
- [Biblical References: Format, Examples,
  History](https://catholic-resources.org/Bible/Biblical_References.htm):
  compiled by Felix Just, S.J., Ph.D.
- [Basic Texts for the Roman Catholic Eucharist - THE ORDER OF
  MASS](https://catholic-resources.org/ChurchDocs/Mass-RM3.htm): compiled by
  Felix Just, S.J., Ph.D.
- [Liturgical Ordo 2023 –
  2024](https://www.liturgyoffice.org.uk/Calendar/2024/index.shtml): from
  Liturgical Office England & Wales.
- [Liturgical Calendar for the Dioceses of the United States of
  America](https://www.usccb.org/committees/divine-worship/liturgical-calendar).
