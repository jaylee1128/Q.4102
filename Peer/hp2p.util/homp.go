//
// The MIT License
//
// Copyright (c) 2022 ETRI
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
//

package util

import (
	"encoding/json"

	"github.com/go-resty/resty/v2"
)

type HOMP struct {
	OverlayAddr string
}

func (self *HOMP) CreateOverlay(hoc *HybridOverlayCreation) *HybridOverlayCreationResponseOverlayInfo {
	PrintJson(WORK, "Create overlay", hoc)

	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(hoc).
		Post(self.OverlayAddr + "/homs")

	if err != nil {
		Println(ERROR, "Create overlay error : ", err)
		return nil
	}

	if resp.StatusCode() != 200 {
		Println(ERROR, "Create overlay error : ", resp.Status(), resp)
		return nil
	}

	ovinfo := new(HybridOverlayCreationResponse)
	json.Unmarshal(resp.Body(), ovinfo)

	PrintJson(WORK, "Create overlay resp", ovinfo)

	return &ovinfo.OverlayInfo
}

func (self *HOMP) QueryOverlay(ovid *string, title *string, desc *string) *HybridOverlayQueryResponse {
	Printf(WORK, "Query Overlay : Overlay ID:%v, Title:%v, Description:%v", ovid, title, desc)

	params := map[string]string{}

	if ovid != nil && len(*ovid) > 0 {
		params["overlay-id"] = *ovid
	}

	if title != nil && len(*title) > 0 {
		params["title"] = *title
	}

	if desc != nil && len(*desc) > 0 {
		params["description"] = *desc
	}

	client := resty.New()
	resp, err := client.R().
		SetQueryParams(params).
		Get(self.OverlayAddr + "/homs")

	if err != nil {
		Println(ERROR, "Query overlay error : ", err)
		return nil
	}

	rslt := new(HybridOverlayQueryResponse)
	json.Unmarshal(resp.Body(), rslt)
	PrintJson(WORK, "Query overlay resp", rslt)

	if len(rslt.Overlay) > 0 {
		for _, overlay := range rslt.Overlay {
			self.convertChannelAttribute(&overlay.ServiceInfo.ChannelList)
		}
	}

	return rslt
}

func (self *HOMP) convertChannelAttribute(channelList *[]ChannelInfo) {
	if channelList != nil && len(*channelList) > 0 {
		for _, channel := range *channelList {
			if channel.ChannelType == "video/feature" && channel.ChannelAttribute != nil {
				cavf := new(ChannelAttributeVideoFeature)
				cai := (*channel.ChannelAttribute).(map[string]interface{})
				cavf.Target = cai["target"].(string)
				cavf.Mode = cai["mode"].(string)
				cavf.Resolution = cai["resolution"].(string)
				cavf.FrameRate = cai["frame-rate"].(string)
				cavf.KeypointsType = cai["keypoints-type"].(string)
				cavf.Dimension = cai["dimension"].(string)
				cavf.ModelUri = cai["model-uri"].(string)
				cavf.Hash = cai["hash"].(string)
				(*channel.ChannelAttribute) = *cavf
			} else if channel.ChannelType == "audio" {
				caa := new(ChannelAttributeAudio)
				cai := (*channel.ChannelAttribute).(map[string]interface{})
				caa.Bitrate = cai["bitrate"].(string)
				caa.SampleRate = cai["sample-rate"].(string)
				caa.ChannelType = cai["channel-type"].(string)
				caa.Codec = cai["codec"].(string)
				(*channel.ChannelAttribute) = *caa
			} else if channel.ChannelType == "text" {
				cat := new(ChannelAttributeText)
				cai := (*channel.ChannelAttribute).(map[string]interface{})
				cat.Encoding = cai["encoding"].(string)
				cat.Format = cai["format"].(string)
				(*channel.ChannelAttribute) = *cat
			}
		}
	}
}

func (self *HOMP) OverlayJoin(hoj *HybridOverlayJoin, recovery bool) *HybridOverlayJoinResponse {

	PrintJson(WORK, "Join overlay", hoj)

	url := "/peer"

	if recovery {
		url += "?recovery=true"
	}

	client := resty.New()
	resp, err := client.R().
		SetBody(hoj).
		Post(self.OverlayAddr + url)

	if err != nil {
		Println(WORK, "Join overlay error : ", err)
		return nil
	}

	ovinfo := new(HybridOverlayJoinResponse)
	json.Unmarshal(resp.Body(), ovinfo)
	ovinfo.RspCode = resp.StatusCode()

	self.convertChannelAttribute(&ovinfo.Overlay.ServiceInfo.ChannelList)

	PrintJson(WORK, "Join overlay resp", ovinfo)

	return ovinfo
}

func (self *HOMP) OverlayModification(hom *HybridOverlayModification) *HybridOverlayModification {

	PrintJson(WORK, "Modification overlay", hom)

	client := resty.New()
	resp, err := client.R().
		SetBody(hom).
		Put(self.OverlayAddr + "/homs")

	if err != nil {
		Println(WORK, "Modification overlay error : ", err)
		return nil
	}

	ovinfo := new(HybridOverlayModification)
	json.Unmarshal(resp.Body(), ovinfo)
	ovinfo.RspCode = resp.StatusCode()

	PrintJson(WORK, "Modification overlay resp", ovinfo)

	return ovinfo
}

func (self *HOMP) OverlayRemoval(hor *HybridOverlayRemoval) *HybridOverlayRemovalResponse {

	PrintJson(WORK, "Remove overlay", hor)

	client := resty.New()
	resp, err := client.R().
		SetBody(hor).
		Delete(self.OverlayAddr + "/homs")

	if err != nil {
		Println(WORK, "Remove overlay error : ", err)
		return nil
	}

	rslt := new(HybridOverlayRemovalResponse)
	json.Unmarshal(resp.Body(), rslt)
	rslt.RspCode = resp.StatusCode()

	PrintJson(WORK, "Remove overlay resp", rslt)

	return rslt
}

func (self *HOMP) OverlayReport(hor *HybridOverlayReport) *HybridOverlayReportOverlay {

	PrintJson(WORK, "Report Overlay", hor)

	client := resty.New()
	resp, err := client.R().
		SetBody(hor).
		Post(self.OverlayAddr + "/peer/report")

	if err != nil {
		Println(WORK, "Report overlay error : ", err)
		return nil
	}

	ovinfo := new(HybridOverlayReportResponse)
	json.Unmarshal(resp.Body(), ovinfo)

	PrintJson(WORK, "Report Overlay resp", ovinfo)

	return &ovinfo.Overlay
}

func (self *HOMP) OverlayRefresh(hor *HybridOverlayRefresh) *HybridOverlayRefreshResponse {

	//self.printJson(hor, "Refresh Overlay: ", INFO)
	//Println(WORK, "Send Refresh Overlay")

	client := resty.New()
	resp, err := client.R().
		SetBody(hor).
		Put(self.OverlayAddr + "/peer")

	if err != nil {
		Println(ERROR, "Refresh overlay error : ", err)
		return nil
	}

	ovinfo := new(HybridOverlayRefreshResponse)
	json.Unmarshal(resp.Body(), ovinfo)

	//self.printJson(ovinfo, "Refresh overlay resp: ", INFO)
	//Println(WORK, "Recv Refresh Overlay resp")

	return ovinfo
}

func (self *HOMP) OverlayLeave(hol *HybridOverlayLeave) *HybridOverlayLeaveResponse {
	Println(WORK, "Send Overlay leave")

	client := resty.New()
	resp, err := client.R().
		SetBody(hol).
		Delete(self.OverlayAddr + "/peer")

	if err != nil {
		Println(ERROR, "Overlay leave error : ", err)
		return nil
	}

	ovl := new(HybridOverlayLeaveResponse)
	json.Unmarshal(resp.Body(), ovl)
	ovl.RspCode = resp.StatusCode()

	Println(WORK, "Recv Overlay leave resp")

	return ovl
}

func (self *HOMP) UserQuery(hou *HybridOverlayUserQuery) *HybridOverlayUserQueryResponse {
	Println(WORK, "Send User Query")

	client := resty.New()
	resp, err := client.R().
		Get(self.OverlayAddr + "/peer?" + "overlay-id=" + hou.OverlayId + "&peer-id=" + hou.PeerId)

	if err != nil {
		Println(ERROR, "User Query error : ", err)
		return nil
	}

	upr := new(HybridOverlayUserQueryResponse)
	json.Unmarshal(resp.Body(), upr)
	upr.RspCode = resp.StatusCode()

	Println(WORK, "Recv User Query resp")

	return upr
}
