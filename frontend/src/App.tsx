import React from 'react';
import { container } from './08_di/container';
import { useRadioPresenter } from './06_adapter_presenter/useRadioPresenter';
import { RadioRoomView } from './07_presentation_view/RadioRoomView';

export const App: React.FC = () => {
  const presenter = useRadioPresenter(container.radioUseCase);

  return (
    <RadioRoomView
      isBroadcasting={presenter.isBroadcasting}
      programTitle={presenter.programTitle}
      currentTheme={presenter.currentTheme}
      isStarting={presenter.isStarting}
      dialogues={presenter.dialogues}
      comments={presenter.comments}
      userName={presenter.userName}
      content={presenter.content}
      isSubmitting={presenter.isSubmitting}
      errorMsg={presenter.errorMsg}
      isPlaying={presenter.isPlaying}
      needsAudioPermission={presenter.needsAudioPermission}
      onProgramTitleChange={presenter.setProgramTitle}
      onUserNameChange={presenter.setUserName}
      onContentChange={presenter.setContent}
      onStartBroadcast={presenter.handleStartBroadcast}
      onSubmitComment={presenter.handlePostComment}
      onEnableAudio={presenter.enableAudio}
    />
  );
};

export default App;
