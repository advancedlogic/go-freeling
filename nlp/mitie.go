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
   char * ret = malloc(strlen(str1)+strlen(str2));

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
	"fmt"
	"unsafe"

	"github.com/abiosoft/semaphore"
)

type Entity struct {
	model string
	prob  float64
	value string
}

func NewEntity(model string, prob float64, value string) *Entity {
	return &Entity{
		model: model,
		prob:  prob,
		value: value,
	}
}

func (this *Entity) String() string {
	return fmt.Sprintf("%s:%0.3f:%s", this.model, this.prob, this.value)
}

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

func (this *MITIE) Process(body string) {
	tokens := C.mitie_tokenize(C.CString(body))
	defer C.mitie_free(unsafe.Pointer(tokens))
	dets := C.mitie_extract_entities(this.ner, tokens)
	defer C.mitie_free(unsafe.Pointer(dets))
	num_dets := C.mitie_ner_get_num_detections(dets)
	for i := 0; i < int(num_dets); i++ {
		centity := C.get_entity(tokens, dets, C.ulong(i))
		entity := NewEntity(C.GoString(centity.model), float64(centity.score), C.GoString(centity.value))
		println(entity.String())
	}
}
