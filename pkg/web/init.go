// Copyright © 2018 Alfred Chou <unioverlord@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package web

import "mime"

func init() {
	mime.AddExtensionType(".html", "text/html; charset=utf-8")
	mime.AddExtensionType(".htm", "text/html; charset=utf-8")
	mime.AddExtensionType(".xml", "text/xml; charset=utf-8")
	mime.AddExtensionType(".js", "text/javascript; charset=utf-8")
	mime.AddExtensionType(".css", "text/css; charset=utf-8")
	mime.AddExtensionType(".txt", "text/plain; charset=utf-8")
	mime.AddExtensionType(".csv", "text/csv; charset=utf-8")
	mime.AddExtensionType(".exe", "application/octet-stream")
	mime.AddExtensionType(".pdf", "application/pdf")
	mime.AddExtensionType(".doc", "application/msword")
	mime.AddExtensionType(".xls", "application/vnd.ms-excel")
	mime.AddExtensionType(".ppt", "application/vnd.ms-powerpoint")
	mime.AddExtensionType(".docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	mime.AddExtensionType(".xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	mime.AddExtensionType(".pptx", "application/vnd.openxmlformats-officedocument.presentationml.presentation")
	mime.AddExtensionType(".jpg", "image/jpeg")
	mime.AddExtensionType(".jpeg", "image/jpeg")
	mime.AddExtensionType(".png", "image/png")
	mime.AddExtensionType(".gif", "image/gif")
	mime.AddExtensionType(".bmp", "image/bmp")
	mime.AddExtensionType(".ico", "image/x-icon")
	mime.AddExtensionType(".zip", "application/zip")
	mime.AddExtensionType(".tar", "application/tar")
	mime.AddExtensionType(".gz", "application/gzip")
	mime.AddExtensionType(".tgz", "application/tar+gzip")
	mime.AddExtensionType(".mp3", "audio/mpeg")
	mime.AddExtensionType(".m4a", "audio/aac")
	mime.AddExtensionType(".ogg", "audio/ogg")
	mime.AddExtensionType(".wav", "audio/wav")
	mime.AddExtensionType(".mpg", "video/mpeg")
	mime.AddExtensionType(".mpeg", "video/mpeg")
	mime.AddExtensionType(".webm", "video/webm")
	mime.AddExtensionType(".mp4", "video/mp4")
	mime.AddExtensionType(".avi", "video/avi")
}
