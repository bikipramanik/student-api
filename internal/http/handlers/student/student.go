package student

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/bikipramanik/students-api/internal/storage"
	"github.com/bikipramanik/students-api/internal/types"
	"github.com/bikipramanik/students-api/internal/utils/response"
	"github.com/go-playground/validator/v10"
)

func New(strg storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var student types.Student

		err := json.NewDecoder(r.Body).Decode(&student)

		if errors.Is(err, io.EOF) {
			slog.Info("EFO error occures")

			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("empty body")))
			return
		}

		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		//request validation

		if err := validator.New().Struct(student); err != nil {

			validateErrs := err.(validator.ValidationErrors)

			response.WriteJson(w, http.StatusBadRequest, response.ValidationError(validateErrs))

			return
		}

		slog.Info("Creating a student")

		lastId, err := strg.CreateStudent(
			student.Name,
			student.Email,
			student.Age,
		)

		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		slog.Info("user created successfully", slog.String("userId", fmt.Sprint(lastId)))

		response.WriteJson(w, http.StatusCreated, map[string]int64{"id": lastId})
	}
}

func GetList(strg storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("getting all students")

		students, err := strg.GetStudents()

		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, err)
			return
		}
		response.WriteJson(w, http.StatusOK, students)
	}
}

func GetById(strg storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		slog.Info("Getting a student", slog.String("id", id))

		intId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			slog.Error("error converting int64 from string", slog.String("id", id))

			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		student, err := strg.GetStudentById(intId)

		if err != nil {
			slog.Error("error getting user", slog.String("id", id))
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}
		response.WriteJson(w, http.StatusOK, student)
	}
}

func DeleteStudent(strg storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		slog.Info("Deleting a student", slog.String("id", id))

		intId, err := strconv.ParseInt(id, 10, 64)

		if err != nil {
			slog.Error("error converting int64 from string", slog.String("id", id))

			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}
		err = strg.DeleteStudentById(intId)

		if err != nil {
			slog.Error("error deleting student", slog.String("id", id))
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}
		response.WriteJson(w, http.StatusOK, map[string]string{"message": "student deleted successfully"})

	}
}

func UpdateStudent(strg storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var students types.Student

		err := json.NewDecoder(r.Body).Decode(&students)

		if errors.Is(err, io.EOF) {
			slog.Info("EFO error occurs")

			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("empty body")))
			return
		}

		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		id := r.PathValue("id")
		slog.Info("Updating a student", slog.String("id", id))

		intId, err := strconv.ParseInt(id, 10, 64)

		if err != nil {
			slog.Error("error converting int64 from string", slog.String("id", id))

			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}
		err = strg.UpdateStudentById(intId, students.Name, students.Email, students.Age)
		if err != nil {
			slog.Error("error updating student", slog.String("id", id))
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return

		}
		students.Id = intId
		response.WriteJson(w, http.StatusOK, response.Success("Student updated successfully", students))
	}
}

func PatchStudent(strg storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PatchStudentRequest

		err := json.NewDecoder(r.Body).Decode(&req)
		if errors.Is(err, io.EOF) {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("empty body")))
			return
		}
		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		if req.Name == nil && req.Email == nil && req.Age == nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("at least one field (name, email, age) must be provided")))
			return
		}

		id := r.PathValue("id")
		slog.Info("Patching a student", slog.String("id", id))

		intId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			slog.Error("error converting int64 from string", slog.String("id", id))
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		err = strg.PatchStudentById(intId, req.Name, req.Email, req.Age)
		if err != nil {
			slog.Error("error patching student", slog.String("id", id), slog.String("error", err.Error()))
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		updatedStudent, err := strg.GetStudentById(intId)
		if err != nil {
			slog.Error("error fetching updated student", slog.String("id", id))
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		response.WriteJson(w, http.StatusOK, response.Success("Student patched successfully", updatedStudent))
	}
}

