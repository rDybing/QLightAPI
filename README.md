# qlight.go

API to handle requests from qlight client and controller(s).

All endpoints requiring client or controller API ID and Key in header.

- [ip]/postAppInfo/ - Registers appID with the API, collects some basic 
information of client/controller such as device OS and attributes such as 
display size.
- [ip]/getServerIP/ - Returns nearest controller (if any) based upon IP
of clients IP. Assumes same public-facing IP of client and controller.
- [ip]/getWelcome/ - downloads a random opening messages.

## Build

Contains the following non-standard libraries:

- github.com/goji/httpauth
- github.com/gorilla/mux

**Contact:**

location   | name/handle | checks
-----------|-------------|--------
github:    | rDybing     | often
Linked In: | Roy Dybing  | monthly
MeWe:      | Roy Dybing  | weekly

---

## Releases

- Version format: [major release].[new feature(s)].[bugfix patch-version]
- Date format: yyyy-mm-dd

#### v.1.0.0: TBA medio August 2019

- First release 

---

## Copyright 2019 Roy Dybing  - all rights reserved.

Source is open to provide insight into working app, mainly to ensure any and 
all that this app do not collect any data of use or user or device it is 
installed upon - except as explicitly noted below:

- ip of device when connecting to WAN API server
- os version and device resolution of client/controller when connecting to WAN 
API server 
- appId of app when connecting to LAN server and WAN API server
- time of contact with WAN API server

Source is not to be used to facilitate distribution of compiled code by any 
third party.

Configuration files and media files are *NOT* included in this source repo.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR 
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, 
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE 
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER 
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, 
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE 
SOFTWARE.

---

ʕ◔ϖ◔ʔ
