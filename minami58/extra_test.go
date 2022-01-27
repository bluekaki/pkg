package minami58

import (
	"bytes"
	"testing"
)

func TestDesc(t *testing.T) {
	payload := []byte(`Lorem ipsum dolor sit amet, et adhuc definitionem his, scripta efficiantur ullamcorper per et. Vim lobortis mediocrem in. Duo ut tantas appetere euripidis, ei est iusto albucius, id dicant sapientem sit. Duo eius atomorum te, usu impetus eligendi et, ei sit meliore perfecto.
Vocent luptatum in mel, sea ex mazim discere. Vel lucilius oportere et. Per eu modus partem sensibus. Idque partem corrumpit ea pri.
Has fierent omnesque qualisque id. Ius id zril cotidieque, ferri consulatu interpretaris eu mei, pro prima eripuit sadipscing an. Eu est aliquid assentior persequeris, quem graecis in pro, primis facete usu no. Nam an dico choro maiorum, sit et admodum platonem gloriatur, quo in patrioque interesset. Ius justo altera nonumes eu, eam esse habemus ea, eam harum referrentur eu.
Vim ne alii intellegat, veri volutpat ne sit. Nemore cetero dissentiet et cum, at appareat inciderint per. Vim partiendo gloriatur contentiones et, falli fierent molestie qui in, iriure petentium vulputate eos ea. Usu et latine labores delicata, ut verear maiestatis est. Cum eu veri timeam adipiscing, sumo malis adipisci ut cum, his nonumy aliquip feugait ut. Cu cum erat molestie, vix nibh ponderum mediocritatem eu.
Graecis volutpat in mel, summo fuisset no est. Qui decore homero euismod cu, ei falli utinam vis. Fabulas persequeris qui in. Vero doming has at, eius consequat mel at. Vis diam putant at.`)

	raw, err := EncodeWithDesc(payload, "Lorem ipsum", LF)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(raw))

	desc, raw, err := DecodeWithDesc(raw)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(raw, payload) {
		t.Fatal("decoded payload not match")
	}

	t.Log(desc)
	t.Log(string(raw))
}
