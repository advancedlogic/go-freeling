package nlp

/*
#cgo LDFLAGS: -L/usr/local/lib -lmitie
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "mitie.h"

typedef struct {
	const char* model;
	double score;
	const char* value;
} Entity;

char * my_strcat(const char * str1, const char * str2);

Entity get_entity(char** tokens,
    const mitie_named_entity_detections* dets,
    unsigned long i) {
	Entity entity;

	unsigned long pos, len;

	pos = mitie_ner_get_detection_position(dets, i);
	len = mitie_ner_get_detection_length(dets, i);

	double score = mitie_ner_get_detection_score(dets,i);
	const char* model = mitie_ner_get_detection_tagstr(dets,i);

	entity.model = model;
	entity.score = score;

	const char* value = "";
	while(len > 0)
	{
		value = my_strcat(value, " ");
		value = my_strcat(value, tokens[pos++]);
		len--;
	}
	entity.value = value;
	return entity;
}

char * my_strcat(const char * str1, const char * str2)
{
	char * ret = malloc(strlen(str1)+strlen(str2)+1);

	if(ret!=NULL)
	{
		sprintf(ret, "%s%s", str1, str2);
		return ret;
	}
	return NULL;
}

void releaseTokens(char** tokens) {
	mitie_free(tokens);
}

void releaseDets(mitie_named_entity_detections* dets) {
	mitie_free(dets);
}

*/
import "C"

import (
	"container/list"
	"fmt"
	"unsafe"

	"github.com/abiosoft/semaphore"
	"github.com/advancedlogic/go-freeling/models"
	"github.com/fatih/set"
)

type MITIE struct {
	ner *C.mitie_named_entity_extractor
	sem *semaphore.Semaphore
}

func NewMITIE(filepath string) *MITIE {
	ner := C.mitie_load_named_entity_extractor(C.CString(filepath))
	sem := semaphore.New(4)
	return &MITIE{
		ner: ner,
		sem: sem,
	}
}

func (this *MITIE) Release() {
	C.mitie_free(unsafe.Pointer(this.ner))
}

func (this *MITIE) Process(body string) *list.List {
	tokens := C.mitie_tokenize(C.CString(body))
	if tokens == nil {
		return nil
	}
	defer C.mitie_free(unsafe.Pointer(tokens))
	dets := C.mitie_extract_entities(this.ner, tokens)
	if dets == nil {
		return nil
	}
	defer C.mitie_free(unsafe.Pointer(dets))
	num_dets := C.mitie_ner_get_num_detections(dets)
	duplicates := set.New(set.ThreadSafe).(*set.Set)
	entites := list.New()
	for i := 0; i < int(num_dets); i++ {
		centity := C.get_entity(tokens, dets, C.ulong(i))
		model := C.GoString(centity.model)
		score := float64(centity.score)
		value := C.GoString(centity.value)
		key := fmt.Sprintf("%s:%s", value, model)
		if duplicates.Has(key) {
			continue
		}
		duplicates.Add(key)
		if score > 0.5 {
			entity := models.NewEntity(model, score, value)
			entites.PushBack(entity)
		}
	}
	return entites
}
