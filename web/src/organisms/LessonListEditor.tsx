import React, { useEffect, useState, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import axios from 'axios';
import { useRecoilState } from 'recoil';
import update from 'immutability-helper';
import { Button } from '@mui/material';
import { userState } from '../config/Recoil';
import { User } from '../config/Type';

function LessonListEditor() {
  const { uuid } = useParams();
  const user: User = useRecoilState(userState)[0];
  const [lessons, setLessons] = useState([
    {
      id: 1,
      text: 'Write a cool JS library',
    },
    {
      id: 2,
      text: 'Make it generic enough',
    },
    {
      id: 3,
      text: 'Write README',
    },
    {
      id: 4,
      text: 'Create some examples',
    },
    {
      id: 5,
      text: 'Spam in Twitter and IRC to promote it (note that this element is taller than the others)',
    },
    {
      id: 6,
      text: '???',
    },
    {
      id: 7,
      text: 'PROFIT',
    },
  ]);
  useEffect(() => {
    if (uuid) {
      console.log('useEffect in home is running');
      axios.get(`http://localhost:8080/api/v1/lessons/${uuid}`, {})
        .then((response) => {
          console.log(response.data);
        }, (error) => {
          console.log(error);
        });
    }
  }, []);

  const moveLesson = useCallback((dragIndex: number, hoverIndex: number) => {
    setLessons((prevLessons) => {
      return update(prevLessons, {
        $splice: [
          [dragIndex, 1],
          [hoverIndex, 0, prevLessons[dragIndex]],
        ],
      });
    });
  }, []);

  const renderLesson = useCallback(
    (lesson: { id: number; text: string }, index: number) => {
      return (
        <Card
          key={lesson.id}
          index={index}
          id={lesson.id}
          text={lesson.text}
          moveLesson={moveLesson}
        />
      );
    },
    [],
  );
  const addLesson = () => {
    console.log('addLesson');
  };

  return (
    <div>
      <Button size="small" onClick={addLesson}>Craete New Lesson</Button>
      <div>{lessons.map((lesson, i) => { return renderLesson(lesson, i); })}</div>
    </div>

  );
}

export default LessonListEditor;
